package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Tris20/FairFareFinder/src/backend"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Weather struct {
	Date           string
	AvgDaytimeTemp sql.NullFloat64
	WeatherIcon    string
	GoogleUrl      string
	AvgDaytimeWpi  sql.NullFloat64
}

type Flight struct {
	DestinationCityName  string
	RandomImageURL       string
	PriceCity1           sql.NullFloat64
	UrlCity1             string
	WeatherForecast      []Weather
	AvgWpi               sql.NullFloat64
	BookingUrl           sql.NullString
	BookingPppn          sql.NullFloat64
	FiveNightsFlights    sql.NullFloat64
	DurationMins         sql.NullInt64
	DurationHours        sql.NullInt64
	DurationHoursRounded sql.NullInt64
	DurationHourDotMins  sql.NullFloat64
}

type FlightsData struct {
	SelectedCity1 string
	Flights       []Flight
	MaxWpi        sql.NullFloat64
	MinFlight     sql.NullFloat64
	MinHotel      sql.NullFloat64
	MinFnaf       sql.NullFloat64
}

// Global variables: template, database, session store
var (
	tmpl       *template.Template
	db         *sql.DB
	store      *sessions.CookieStore = sessions.NewCookieStore([]byte("your-secret-key"))
	mutePrints                       = false
)

func setMutePrints(value bool) {
	mutePrints = value
}

func main() {
	// Parse the "web" flag
	webFlag := flag.Bool("web", false, "Pass this flag to enable the web server with file check routine")
	flag.Parse() // Parse command-line flags

	// Create a lumberjack logger
	fileLogger := &lumberjack.Logger{
		Filename:   "./app.log", // File to log to
		MaxSize:    69,          // Maximum size in megabytes before it gets rotated
		MaxBackups: 5,           // Max number of old log files to keep
		MaxAge:     28,          // Max number of days to retain log files
		Compress:   true,        // Compress the rotated files using gzip
	}

	// Set up the server
	// pass in database path and logger for testing purposes
	cleanup := SetupServer("./data/compiled/main.db", fileLogger)
	defer cleanup()

	// On web server, every 2 hours, check for a new database delivery, and swap dbs accordingly
	fmt.Printf("Flag? Value: %v\n", *webFlag)
	if *webFlag {
		fmt.Println("Starting db monitor")
		go backend.StartFileCheckRoutine(&db, &tmpl)
	}

	// Start the server
	StartServer()
}

func SetupServer(db_path string, logger io.Writer) func() {
	// Set up lumberjack log file rotation config
	log.SetOutput(logger)

	var err error

	db, err = sql.Open("sqlite3", db_path)
	if err != nil {
		log.Fatal(err)
	}

	// Load city-country pairs into memory searchbar to use
	backend.LoadCityCountryPairs(db)

	cleanup := func() {
		if db != nil {
			db.Close()
		}
	}
	// Register custom functions for templates
	tmpl = template.Must(template.New("").Funcs(template.FuncMap{
		"toJson": func(v interface{}) (string, error) {
			a, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(a), nil
		},
	}).ParseFiles(
		"./src/frontend/html/index.html",
		"./src/frontend/html/table.html",
		"./src/frontend/html/seo.html",
	))

	backend.Init(db, tmpl)

	// Set up routes
	http.HandleFunc("/", backend.IndexHandler)
	http.HandleFunc("/filter", combinedCardsHandler)
	http.HandleFunc("/update-slider-price", backend.UpdateSliderPriceHandler)

	// Serve static files
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./src/frontend/css/"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./src/frontend/images"))))
	http.Handle("/location-images/", http.StripPrefix("/location-images/", http.FileServer(http.Dir("./ignore/location-images"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./src/frontend/js/")))) // New JS route

	http.HandleFunc("/city-country-pairs", backend.CityCountryHandler)

	// Privacy policy route
	http.HandleFunc("/privacy-policy", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./src/frontend/html/privacy-policy.html") // Ensure the path is correct
	})

	return cleanup
}

func StartServer() {
	// Listen on all network interfaces including localhost
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

func combinedCardsHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	cities := r.URL.Query()["city[]"]
	logicalOperators := r.URL.Query()["logical_operator[]"]
	maxPriceLinearStrs := r.URL.Query()["maxPriceLinear[]"]
	maxAccomPriceLinearStrs := r.URL.Query()["maxAccommodationPrice[]"]

	// Validate input lengths
	if len(cities) == 0 || len(cities) != len(logicalOperators)+1 || len(cities) != len(maxPriceLinearStrs) {
		response := fmt.Sprintf("Mismatched input lengths. Cities: %d, Operators: %d, Prices: %d",
			len(cities), len(logicalOperators), len(maxPriceLinearStrs))
		http.Error(w, response, http.StatusBadRequest)
		return
	}

	// Parse price limits
	var maxPrices []float64
	for _, linearStr := range maxPriceLinearStrs {
		linearValue, err := strconv.ParseFloat(linearStr, 64)
		if err != nil {
			http.Error(w, "Invalid price parameter", http.StatusBadRequest)
			return
		}
		maxPrices = append(maxPrices, backend.MapLinearToExponential(linearValue, 50, 2500))
	}

	// Parse logical expression
	expr, err := parseLogicalExpression(cities, logicalOperators, maxPrices)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve and process sort option
	sortOption := r.URL.Query().Get("sort")
	if sortOption == "" {
		sortOption = "low_price" // default
	}
	orderClause := determineOrderClause(sortOption)

	// Parse accommodation price limit
	var maxAccommodationPrice float64
	if len(maxAccomPriceLinearStrs) > 0 {
		accomLinearStr := maxAccomPriceLinearStrs[0]
		accomLinearValue, err := strconv.ParseFloat(accomLinearStr, 64)
		if err != nil {
			http.Error(w, "Invalid accommodation price parameter", http.StatusBadRequest)
			return
		}
		maxAccommodationPrice = backend.AccomMapLinearToExponential(accomLinearValue, 10, 550)
	} else {
		// Default value if no accommodation price is provided
		maxAccommodationPrice = 70.0
	}

	// Build the query

	query, args := buildQuery(expr, maxAccommodationPrice, cities, orderClause)

	if !mutePrints {
		// Output the query for debugging
		fmt.Println("Generated SQL Query:")
		fmt.Println(query)
		fmt.Println("Arguments:")
		fmt.Println(args)
	}
	// Log the interpolated query for debugging
	fullQuery := interpolateQuery(query, args)
	log.Printf("Full Query:\n%s\n", fullQuery)

	// Check if db is nil
	if db == nil {
		http.Error(w, "Database connection is not initialized", http.StatusInternalServerError)
		return
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	// Process results
	flights, err := processFlightRows(rows)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Save the session and render the response
	//session.Values["city1"] = cities[0] // Save the first city
	session.Save(r, w)

	data := buildFlightsData(cities, flights)
	err = tmpl.ExecuteTemplate(w, "table.html", data)
	if err != nil {
		http.Error(w, "Error rendering results", http.StatusInternalServerError)
	}
}

// Helper function to process rows into flight and weather data
func processFlightRows(rows *sql.Rows) ([]Flight, error) {
	var flights []Flight
	for rows.Next() {
		var flight Flight
		var weather Weather
		var imageUrl sql.NullString
		var bookingUrl sql.NullString
		var priceFnaf sql.NullFloat64
		var duration_mins sql.NullInt64
		var duration_hours sql.NullInt64
		var duration_hours_rounded sql.NullInt64
		var duration_hour_dot_mins sql.NullFloat64

		err := rows.Scan(
			&flight.DestinationCityName,
			&flight.PriceCity1,
			&flight.UrlCity1,
			&weather.Date,
			&weather.AvgDaytimeTemp,
			&weather.WeatherIcon,
			&weather.GoogleUrl,
			&flight.AvgWpi,
			&imageUrl,
			&bookingUrl,
			&flight.BookingPppn,
			&priceFnaf,
			&duration_mins,
			&duration_hours,
			&duration_hours_rounded,
			&duration_hour_dot_mins,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		if duration_mins.Valid {
			flight.DurationMins = duration_mins
			log.Printf("Duration: %d hours for flight to %s", duration_mins.Int64, flight.DestinationCityName)
		} else {
			log.Printf("No valid duration found for flight to %s", flight.DestinationCityName)
		}
		if duration_hours.Valid {
			flight.DurationHours = duration_hours
			log.Printf("Duration: %d hours for flight to %s", duration_hours.Int64, flight.DestinationCityName)
		} else {
			log.Printf("No valid duration found for flight to %s", flight.DestinationCityName)
		}
		if duration_hours_rounded.Valid {
			flight.DurationHoursRounded = duration_hours_rounded
			log.Printf("Duration: %d hours for flight to %s", duration_hours_rounded.Int64, flight.DestinationCityName)
		} else {
			log.Printf("No valid duration found for flight to %s", flight.DestinationCityName)
		}
		if duration_hour_dot_mins.Valid {
			flight.DurationHourDotMins = duration_hour_dot_mins
			log.Printf("Duration: %.2f hours.mins for flight to %s", duration_hour_dot_mins.Float64, flight.DestinationCityName)
		} else {
			log.Printf("No valid duration found for flight to %s", flight.DestinationCityName)
		}

		// Log the weather data for debugging
		log.Printf("Row Data - Destination: %s, Date: %s, Temp: %.2f, Icon: %s, Duration.Hours: %d, Duration.Mins: %d ",
			flight.DestinationCityName,
			weather.Date,
			weather.AvgDaytimeTemp.Float64,
			weather.WeatherIcon,
			flight.DurationHours.Int64,
			flight.DurationMins.Int64,
		)

		// Log the imageUrl for debugging
		log.Printf("Scanned image URL: '%s', Valid: %t", imageUrl.String, imageUrl.Valid)

		if imageUrl.Valid && len(imageUrl.String) > 5 {
			flight.RandomImageURL = imageUrl.String
			log.Printf("Using image URL from database: %s", flight.RandomImageURL)
		} else {
			flight.RandomImageURL = "/images/location-placeholder-image.png"
			log.Printf("Using default placeholder image URL: %s", flight.RandomImageURL)
		}
		flight.BookingUrl = bookingUrl
		flight.FiveNightsFlights = priceFnaf
		addOrUpdateFlight(&flights, flight, weather)
	}
	return flights, nil
}

// Helper function to add or update flight entries
func addOrUpdateFlight(flights *[]Flight, flight Flight, weather Weather) {
	for i := range *flights {
		if (*flights)[i].DestinationCityName == flight.DestinationCityName {
			(*flights)[i].WeatherForecast = append((*flights)[i].WeatherForecast, weather)
			return
		}
	}

	flight.WeatherForecast = []Weather{weather}
	*flights = append(*flights, flight)
}

// Helper function to build the data for the template
func buildFlightsData(cities []string, flights []Flight) FlightsData {
	// Ensure there is at least one city in the list
	var selectedCity1 string
	if len(cities) > 0 {
		selectedCity1 = cities[0]
	} else {
		selectedCity1 = "" // Default to an empty string if no cities are provided
	}

	// Initialize variables for max/min values
	var maxWpi, minFlightPrice, minHotelPrice, minFnafPrice sql.NullFloat64

	// Process each flight to find max/min values
	for _, flight := range flights {
		maxWpi = backend.UpdateMaxValue(maxWpi, flight.AvgWpi)
		minFlightPrice = backend.UpdateMinValue(minFlightPrice, flight.PriceCity1)
		minHotelPrice = backend.UpdateMinValue(minHotelPrice, flight.BookingPppn)
		minFnafPrice = backend.UpdateMinValue(minFnafPrice, flight.FiveNightsFlights)
	}

	// Build and return the FlightsData
	return FlightsData{
		SelectedCity1: selectedCity1,
		Flights:       flights,
		MaxWpi:        maxWpi,
		MinFlight:     minFlightPrice,
		MinHotel:      minHotelPrice,
		MinFnaf:       minFnafPrice,
	}
}

// Unified Query Builder

func buildQuery(expr Expression, maxAccommodationPrice float64, originCities []string, orderClause string) (string, []interface{}) {
	var queryBuilder strings.Builder
	var args []interface{}
	// Begin the query with the DestinationSet CTE
	queryBuilder.WriteString("WITH DestinationSet AS (\n")

	// Build the subquery based on the logical expression
	subquery, subqueryArgs := buildSubquery(expr)
	queryBuilder.WriteString(subquery)
	queryBuilder.WriteString("\n)")
	args = append(args, subqueryArgs...)

	// Build the rest of the query
	queryBuilder.WriteString(`
    SELECT 
        ds.destination_city_name,
        MIN(f.price_next_week) AS price_city1,
        MIN(f.skyscanner_url_next_week) AS url_city1,
        w.date,
        w.avg_daytime_temp,
        w.weather_icon,
        w.google_url,
        l.avg_wpi,
        l.image_1,
        a.booking_url,
        a.booking_pppn,
        fnf.price_fnaf,
        MIN(f.duration_in_minutes) AS duration_mins,
        MIN(f.duration_in_hours) AS duration_hours,
        MIN(f.duration_in_hours_rounded) AS duration_hours_rounded,
        MIN(f.duration_hour_dot_mins) AS duration_hour_dot_mins
    FROM DestinationSet ds
    JOIN flight f ON ds.destination_city_name = f.destination_city_name 
                   AND ds.destination_country = f.destination_country
    JOIN location l ON ds.destination_city_name = l.city 
                     AND ds.destination_country = l.country
    JOIN weather w ON w.city = ds.destination_city_name 
                    AND w.country = ds.destination_country
    LEFT JOIN accommodation a ON a.city = ds.destination_city_name 
                               AND a.country = ds.destination_country
    LEFT JOIN (
        SELECT 
            fnf.origin_city,
            fnf.origin_country,
            fnf.destination_city,
            fnf.destination_country,
            MIN(fnf.price_fnaf) AS price_fnaf
        FROM five_nights_and_flights fnf
        GROUP BY fnf.origin_city, fnf.origin_country, fnf.destination_city, fnf.destination_country
    ) fnf ON fnf.destination_city = ds.destination_city_name
           AND fnf.destination_country = ds.destination_country
           AND fnf.origin_city = f.origin_city_name
           AND fnf.origin_country = f.origin_country
    WHERE l.avg_wpi BETWEEN 1.0 AND 10.0 
      AND w.date >= date('now')
      AND f.price_next_week < ?
      AND f.origin_city_name IN `)

	// Build the IN clause dynamically based on the number of origin cities
	if len(originCities) == 0 {
		return "", nil // If no origin cities, return an empty query
	}

	placeholders := make([]string, len(originCities))
	for i := range originCities {
		placeholders[i] = "?"
	}
	inClause := fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))
	queryBuilder.WriteString(inClause)
	queryBuilder.WriteString(`
      AND a.booking_pppn IS NOT NULL
      AND a.booking_pppn <= ?
   GROUP BY f.destination_city_name, w.date, f.destination_country, l.avg_wpi
    `)

	// Add the dynamic order clause
	queryBuilder.WriteString(orderClause)
	queryBuilder.WriteString(";")

	// Add price limits and origin city names to args
	maxPrice := 2000.0 // Set a max price or calculate based on inputs
	args = append(args, maxPrice)

	// Correctly append originCities to args
	for _, city := range originCities {
		args = append(args, city)
	}

	args = append(args, maxAccommodationPrice)

	return queryBuilder.String(), args
}

func buildSubquery(expr Expression) (string, []interface{}) {
	switch e := expr.(type) {
	case *CityCondition:
		// Return the subquery for a city condition
		subquery := `
            SELECT 
                f.destination_city_name,
                f.destination_country
            FROM flight f
            WHERE f.origin_city_name = ? AND f.price_next_week < ?
            GROUP BY f.destination_city_name, f.destination_country
        `
		args := []interface{}{e.City.Name, e.City.PriceLimit}
		return subquery, args
	case *LogicalExpression:
		// Build the left and right subqueries
		leftSubquery, leftArgs := buildSubquery(e.Left)
		rightSubquery, rightArgs := buildSubquery(e.Right)
		var operator string
		if e.Operator == AndOperator {
			operator = "INTERSECT"
		} else if e.Operator == OrOperator {
			operator = "UNION"
		} else {
			panic("Unknown operator")
		}
		// Combine subqueries without unnecessary parentheses
		combinedSubquery := fmt.Sprintf("%s\n%s\n%s", leftSubquery, operator, rightSubquery)
		args := append(leftArgs, rightArgs...)
		return combinedSubquery, args
	default:
		panic("Unknown expression type")
	}
}

// Helper function to determine the ORDER BY clause
func determineOrderClause(sortOption string) string {
	switch sortOption {
	case "low_price":
		return "ORDER BY fnf.price_fnaf ASC"
	case "high_price":
		return "ORDER BY fnf.price_fnaf DESC"
	case "best_weather":
		return "ORDER BY avg_wpi DESC"
	case "worst_weather":
		return "ORDER BY avg_wpi ASC"
	case "cheapest_hotel":
		return "ORDER BY a.booking_pppn ASC"
	case "most_expensive_hotel":
		return "ORDER BY a.booking_pppn DESC"
	case "shortest_flight":
		return "ORDER BY f.duration_hour_dot_mins ASC"
	case "longest_flight":
		return "ORDER BY f.duration_hour_dot_mins DESC"
	default:
		return "ORDER BY fnf.price_fnaf ASC" // Default sorting by lowest FNAF price
	}
}

/*
// / Helper to construct SELECT clause
func selectClause() string {
	return `
        SELECT f1.destination_city_name,
               MIN(f1.price_this_week) AS price_city1,
               MIN(f1.skyscanner_url_this_week) AS url_city1,
               w.date,
               w.avg_daytime_temp,
               w.weather_icon,
               w.google_url,
               l.avg_wpi,
               l.image_1,
               a.booking_url,
               a.booking_pppn,
               fnf.price_fnaf
    `
}

// Helper to construct JOIN clause
func joinClause() string {
	return `
        FROM flight f1
        JOIN location l ON f1.destination_city_name = l.city AND f1.destination_country = l.country
        JOIN weather w ON w.city = f1.destination_city_name AND w.country = f1.destination_country
        LEFT JOIN accommodation a ON a.city = f1.destination_city_name AND a.country = f1.destination_country
        LEFT JOIN five_nights_and_flights fnf ON fnf.destination_city = f1.destination_city_name AND fnf.origin_city = ?
    `
}

// Helper to construct WHERE clause

func whereClause(city1 string, additionalCities []string, logicalOperators []string) string {
	// Start with the primary condition for city1
	whereClause := "WHERE f1.origin_city_name = ?"

	if len(additionalCities) > 0 {
		// Use INTERSECT for AND logic, or UNION for OR logic
		subqueries := []string{}
		for i := range additionalCities {
			if i < len(logicalOperators) && logicalOperators[i] == "AND" {
				// Create an INTERSECT query for the additional city
				subqueries = append(subqueries, fmt.Sprintf(`
					SELECT f.destination_city_name
					FROM flight f
					WHERE f.origin_city_name = ?
				`))
			} else if i < len(logicalOperators) && logicalOperators[i] == "OR" {
				// Create a UNION query for the additional city
				subqueries = append(subqueries, fmt.Sprintf(`
					SELECT f.destination_city_name
					FROM flight f
					WHERE f.origin_city_name = ?
				`))
			} else {
				log.Printf("Warning: Mismatch between additionalCities and logicalOperators at index %d", i)
			}
		}

		// Combine subqueries into the WHERE clause
		if len(subqueries) > 0 {
			whereClause += " AND f1.destination_city_name IN ("
			if logicalOperators[0] == "AND" {
				whereClause += strings.Join(subqueries, " INTERSECT ")
			} else {
				whereClause += strings.Join(subqueries, " UNION ")
			}
			whereClause += ")"
		}
	}

	// Add static conditions
	whereClause += `
        AND l.avg_wpi BETWEEN ? AND ?
        AND w.date >= date('now')
    `
	log.Printf("Generated WHERE Clause: %s", whereClause)
	return whereClause
}

// Helper to construct GROUP BY clause
func groupByClause() string {
	return `
        GROUP BY f1.destination_city_name, w.date, f1.destination_country, l.avg_wpi
    `
}

// Helper to construct HAVING clause
func havingClause() string {
	return `
        HAVING MIN(f1.price_this_week) <= ?
    `
}
*/
/*---------------Logical Expressions-----------------------*/

// CityInput represents the input for each city
type CityInput struct {
	Name       string
	PriceLimit float64
}

// LogicalOperator represents a logical operator (AND, OR)
type LogicalOperator string

const (
	AndOperator LogicalOperator = "AND"
	OrOperator  LogicalOperator = "OR"
)

// Expression represents a logical expression
type Expression interface{}

// CityCondition represents a condition for a single city
type CityCondition struct {
	City CityInput
}

// LogicalExpression represents a logical combination of expressions
type LogicalExpression struct {
	Operator LogicalOperator
	Left     Expression
	Right    Expression
}

func parseLogicalExpression(cities []string, logicalOperators []string, maxPrices []float64) (Expression, error) {
	// Validate input lengths
	if len(cities) == 0 || len(cities) != len(maxPrices) || len(cities) != len(logicalOperators)+1 {
		return nil, fmt.Errorf("mismatched input lengths")
	}

	// Base case: Only one city
	if len(cities) == 1 {
		log.Printf("parseLogicalExpression: Single city: %s, PriceLimit: %.2f", cities[0], maxPrices[0])
		return &CityCondition{
			City: CityInput{Name: cities[0], PriceLimit: maxPrices[0]},
		}, nil
	}

	// Start with the first city as the base expression
	log.Printf("parseLogicalExpression: Starting with city: %s, PriceLimit: %.2f", cities[0], maxPrices[0])
	var expr Expression = &CityCondition{
		City: CityInput{Name: cities[0], PriceLimit: maxPrices[0]},
	}

	// Process subsequent cities with their logical operators
	for i := 1; i < len(cities); i++ {
		log.Printf("parseLogicalExpression: Adding city: %s, PriceLimit: %.2f with Operator: %s", cities[i], maxPrices[i], logicalOperators[i-1])
		expr = &LogicalExpression{
			Operator: LogicalOperator(logicalOperators[i-1]),
			Left:     expr,
			Right: &CityCondition{
				City: CityInput{Name: cities[i], PriceLimit: maxPrices[i]},
			},
		}
	}

	return expr, nil
}

func interpolateQuery(query string, args []interface{}) string {
	var result strings.Builder
	argIndex := 0

	for _, char := range query {
		if char == '?' && argIndex < len(args) {
			// Append the argument in place of the '?'
			arg := args[argIndex]
			argIndex++

			// Format the argument based on its type
			switch v := arg.(type) {
			case string:
				result.WriteString(fmt.Sprintf("'%s'", v)) // Quote strings
			case float64:
				result.WriteString(fmt.Sprintf("%.2f", v))
			case int:
				result.WriteString(fmt.Sprintf("%d", v))
			default:
				result.WriteString(fmt.Sprintf("%v", v)) // Fallback for other types
			}
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}
