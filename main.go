package main

import (
	// Standard Library
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	// Third-Party Packages
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/natefinch/lumberjack.v2"

	// Local Packages
	"github.com/Tris20/FairFareFinder/src/backend"
	"github.com/Tris20/FairFareFinder/src/backend/dev_tools"
	"github.com/Tris20/FairFareFinder/src/backend/model"
)

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

	// Initialize templates
	tmpl, err = backend.InitializeTemplates()
	if err != nil {
		log.Fatalf("Failed to initialize templates: %v", err)
	}

	backend.Init(db, tmpl)

	// Set up routes
	http.HandleFunc("/", backend.IndexHandler)
	http.HandleFunc("/filter", filterRequestHandler)
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
	http.HandleFunc("/all-cities", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		clientID := session.ID
		dev_tools.AllCitiesHandler(db, tmpl, clientID)(w, r)
	})
	http.HandleFunc("/load-more-cities", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		clientID := session.ID
		dev_tools.LoadMoreCities(tmpl, clientID)(w, r)
	})

	return cleanup
}

func StartServer() {
	// Listen on all network interfaces including localhost
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

/*
#
#
#
*/
type FilterInput struct {
	Cities                []string
	LogicalOperators      []string
	MaxFlightPrices       []float64
	MaxAccommodationPrice float64
	OrderClause           string
	LogicalExpression     Expression
}

func handleHTTPError(w http.ResponseWriter, message string, code int) {
	log.Printf("Error: %s", message)
	http.Error(w, message, code)
}

func getUserSession(r *http.Request) (*sessions.Session, error) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Error retrieving user session: %v", err)
		return nil, err
	}
	return session, nil
}

func parseAndValidateFilterInputs(r *http.Request) (*FilterInput, error) {
	// Extract query parameters
	cities := r.URL.Query()["city[]"]
	logicalOperators := r.URL.Query()["logical_operator[]"]
	maxFlightPriceLinearStrs := r.URL.Query()["maxFlightPriceLinear[]"]
	maxAccomPriceLinearStrs := r.URL.Query()["maxAccommodationPrice[]"]
	sortOption := r.URL.Query().Get("sort")

	// Validate input lengths
	if len(cities) == 0 || len(cities) != len(logicalOperators)+1 || len(cities) != len(maxFlightPriceLinearStrs) {
		return nil, fmt.Errorf("mismatched input lengths. Cities: %d, Operators: %d, Prices: %d",
			len(cities), len(logicalOperators), len(maxFlightPriceLinearStrs))
	}

	// Parse flight price limits
	maxFlightPrices := make([]float64, 0, len(maxFlightPriceLinearStrs))
	for _, linearStr := range maxFlightPriceLinearStrs {
		linearValue, err := strconv.ParseFloat(linearStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid flight price parameter")
		}
		mappedValue := backend.MapLinearToExponential(linearValue, 50, 1000, 2500)
		maxFlightPrices = append(maxFlightPrices, mappedValue)
	}

	// Parse logical expression
	expr, err := parseLogicalExpression(cities, logicalOperators, maxFlightPrices)
	if err != nil {
		return nil, err
	}

	// Determine sort option and order clause
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
			return nil, fmt.Errorf("invalid accommodation price parameter")
		}
		maxAccommodationPrice = backend.MapLinearToExponential(accomLinearValue, 10, 200, 550)
	} else {
		maxAccommodationPrice = 70.0 // Default value
	}

	return &FilterInput{
		Cities:                cities,
		LogicalOperators:      logicalOperators,
		MaxFlightPrices:       maxFlightPrices,
		MaxAccommodationPrice: maxAccommodationPrice,
		OrderClause:           orderClause,
		LogicalExpression:     expr,
	}, nil
}

func executeMainQuery(input *FilterInput) ([]model.Flight, error) {
	query, args := buildQuery(input.LogicalExpression, input.MaxAccommodationPrice, input.Cities, input.OrderClause)

	if !mutePrints {
		fmt.Println("Generated SQL Query (MAIN):")
		fmt.Println(query)
		fmt.Println("Arguments:", args)
	}

	fullQuery := interpolateQuery(query, args)
	log.Printf("Full MAIN Query:\n%s\n", fullQuery)

	if db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error querying main results: %v", err)
		return nil, err
	}
	defer rows.Close()

	flights, err := processFlightRows(rows)
	if err != nil {
		log.Printf("Error processing flight rows: %v", err)
		return nil, err
	}

	return flights, nil
}

func executeAccommodationPricesHistogramQuery(input *FilterInput) ([]model.Flight, error) {
	// Build query with fixed maxAccommodationPrice = 550.0
	allPricesQuery, allPricesArgs := buildQuery(input.LogicalExpression, 550.0, input.Cities, input.OrderClause)

	if !mutePrints {
		fmt.Println("Generated SQL Query (ALL PRICES):")
		fmt.Println(allPricesQuery)
		fmt.Println("Arguments:", allPricesArgs)
	}

	fullAllPricesQuery := interpolateQuery(allPricesQuery, allPricesArgs)
	log.Printf("Full ALL-PRICES Query:\n%s\n", fullAllPricesQuery)

	if db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	rows, err := db.Query(allPricesQuery, allPricesArgs...)
	if err != nil {
		log.Printf("Error querying all accommodation prices: %v", err)
		return nil, err
	}
	defer rows.Close()

	flightsAll, err := processFlightRows(rows)
	if err != nil {
		log.Printf("Error processing flight rows for all accommodation prices: %v", err)
		return nil, err
	}

	return flightsAll, nil
}

func buildTemplateData(cities []string, flights []model.Flight, allAccomPrices []float64) model.FlightsData {
	data := buildFlightsData(cities, flights)
	data.AllAccommodationPrices = allAccomPrices
	return data
}

func filterRequestHandler(w http.ResponseWriter, r *http.Request) {
	//  Session Management
	session, err := getUserSession(r)
	if err != nil {
		handleHTTPError(w, "Session retrieval error", http.StatusInternalServerError)
		return
	}

	//  Input Extraction and Validation
	input, err := parseAndValidateFilterInputs(r)
	if err != nil {
		handleHTTPError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Execute Main Query to Populate Destination Cards
	flights, err := executeMainQuery(input)
	if err != nil {
		handleHTTPError(w, "Error executing main query", http.StatusInternalServerError)
		return
	}

	//  Execute Second Query to Populate Accommodation Price Slider Histogram
	flightsAll, err := executeAccommodationPricesHistogramQuery(input)
	if err != nil {
		handleHTTPError(w, "Error executing all prices query", http.StatusInternalServerError)
		return
	}

	//  Collect Accommodation Prices
	allAccomPrices := gatherBookingPppn(flightsAll)
	log.Printf("All accommodation prices (no user limit): %v", allAccomPrices)

	//  Prepare Data for the Template
	data := buildTemplateData(input.Cities, flights, allAccomPrices)

	// Save Session and Render the Response
	if err := session.Save(r, w); err != nil {
		handleHTTPError(w, "Session save error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "table.html", data); err != nil {
		handleHTTPError(w, "Error rendering results", http.StatusInternalServerError)
		return
	}
}

/*
#
#
#
*/
// Helper function to process rows into flight and weather data
func processFlightRows(rows *sql.Rows) ([]model.Flight, error) {
	var flights []model.Flight
	for rows.Next() {
		var flight model.Flight
		var weather model.Weather
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

		backend.SetFlightDurationInt(&flight, duration_mins, &flight.DurationMins, "Duration: %d minutes for flight to %s")
		backend.SetFlightDurationInt(&flight, duration_hours, &flight.DurationHours, "Duration: %d hours for flight to %s")
		backend.SetFlightDurationInt(&flight, duration_hours_rounded, &flight.DurationHoursRounded, "Duration: %d rounded hours for flight to %s")
		backend.SetFlightDurationFloat(&flight, duration_hour_dot_mins, &flight.DurationHourDotMins, "Duration: %.2f hours.mins for flight to %s")

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
func addOrUpdateFlight(flights *[]model.Flight, flight model.Flight, weather model.Weather) {
	for i := range *flights {
		if (*flights)[i].DestinationCityName == flight.DestinationCityName {
			(*flights)[i].WeatherForecast = append((*flights)[i].WeatherForecast, weather)
			return
		}
	}

	flight.WeatherForecast = []model.Weather{weather}
	*flights = append(*flights, flight)
}

// Helper function to build the data for the template
func buildFlightsData(cities []string, flights []model.Flight) model.FlightsData {
	// Ensure there is at least one city in the list
	var selectedCity1 string
	if len(cities) > 0 {
		selectedCity1 = cities[0]
	} else {
		selectedCity1 = "" // Default to an empty string if no cities are provided
	}

	// Initialize variables for max/min values
	var maxWpi, minFlightPrice, minHotelPrice, minFnafPrice sql.NullFloat64

	// Collect *all* booking_pppn into a slice
	var allAccomPrices []float64

	// Process each flight to find max/min values
	for _, flight := range flights {
		maxWpi = backend.UpdateMaxValue(maxWpi, flight.AvgWpi)
		minFlightPrice = backend.UpdateMinValue(minFlightPrice, flight.PriceCity1)
		minHotelPrice = backend.UpdateMinValue(minHotelPrice, flight.BookingPppn)
		minFnafPrice = backend.UpdateMinValue(minFnafPrice, flight.FiveNightsFlights)
		if flight.BookingPppn.Valid {
			allAccomPrices = append(allAccomPrices, flight.BookingPppn.Float64)
		}
	}

	// Build and return the FlightsData
	return model.FlightsData{
		SelectedCity1:          selectedCity1,
		Flights:                flights,
		MaxWpi:                 maxWpi,
		MinFlight:              minFlightPrice,
		MinHotel:               minHotelPrice,
		MinFnaf:                minFnafPrice,
		AllAccommodationPrices: allAccomPrices,
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

var orderByClauses = map[string]string{
	"low_price":            "ORDER BY fnf.price_fnaf ASC",
	"high_price":           "ORDER BY fnf.price_fnaf DESC",
	"best_weather":         "ORDER BY avg_wpi DESC",
	"worst_weather":        "ORDER BY avg_wpi ASC",
	"cheapest_hotel":       "ORDER BY a.booking_pppn ASC",
	"most_expensive_hotel": "ORDER BY a.booking_pppn DESC",
	"shortest_flight":      "ORDER BY f.duration_hour_dot_mins ASC",
	"longest_flight":       "ORDER BY f.duration_hour_dot_mins DESC",
}

func determineOrderClause(sortOption string) string {
	if clause, found := orderByClauses[sortOption]; found {
		return clause
	}
	return "ORDER BY fnf.price_fnaf ASC" // Default
}

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

// gatherBookingPppn just extracts all booking_pppn values from flights
func gatherBookingPppn(flights []model.Flight) []float64 {
	var prices []float64
	for _, f := range flights {
		if f.BookingPppn.Valid {
			prices = append(prices, f.BookingPppn.Float64)
		}
	}
	return prices
}
