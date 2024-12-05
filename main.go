package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"

	//"math/rand"
	"net/http"
	//"os"
	//"path/filepath"
	"strconv"
	//"time"
	"github.com/Tris20/FairFareFinder/src/backend"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/natefinch/lumberjack.v2"
	"strings"
)

type Weather struct {
	Date           string
	AvgDaytimeTemp sql.NullFloat64
	WeatherIcon    string
	GoogleUrl      string
	AvgDaytimeWpi  sql.NullFloat64
}

type Flight struct {
	DestinationCityName string
	RandomImageURL      string
	PriceCity1          sql.NullFloat64
	UrlCity1            string
	WeatherForecast     []Weather
	AvgWpi              sql.NullFloat64
	BookingUrl          sql.NullString
	BookingPppn         sql.NullFloat64
	FiveNightsFlights   sql.NullFloat64
}

type FlightsData struct {
	SelectedCity1 string
	Flights       []Flight
	MaxWpi        sql.NullFloat64
	MinFlight     sql.NullFloat64
	MinHotel      sql.NullFloat64
	MinFnaf       sql.NullFloat64
}

var (
	tmpl  *template.Template
	db    *sql.DB
	store *sessions.CookieStore = sessions.NewCookieStore([]byte("your-secret-key"))
)

func main() {
	// Set up lumberjack log file rotation config
	log.SetOutput(&lumberjack.Logger{
		Filename:   "./app.log", // File to log to
		MaxSize:    69,          // Maximum size in megabytes before it gets rotated
		MaxBackups: 5,           // Max number of old log files to keep
		MaxAge:     28,          // Max number of days to retain log files
		Compress:   true,        // Compress the rotated files using gzip
	})

	// Parse the "web" flag
	webFlag := flag.Bool("web", false, "Pass this flag to enable the web server with file check routine")
	flag.Parse() // Parse command-line flags

	var err error

	db, err = sql.Open("sqlite3", "./data/compiled/main.db")

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tmpl = template.Must(template.ParseFiles(
		"./src/frontend/html/index.html",
		"./src/frontend/html/table.html"))

	backend.Init(db, tmpl)

	// Set up routes
	http.HandleFunc("/", backend.IndexHandler)
	http.HandleFunc("/filter", combinedCardsHandler)
	http.HandleFunc("/update-slider-price", backend.UpdateSliderPriceHandler)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./src/frontend/css/"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./src/frontend/images"))))
	http.Handle("/location-images/", http.StripPrefix("/location-images/", http.FileServer(http.Dir("./ignore/location-images"))))
	// Privacy policy route
	http.HandleFunc("/privacy-policy", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./src/frontend/html/privacy-policy.html") // Make sure the path is correct
	})

	// On web server, every 2 hours, check for a new database delivery, and swap dbs accordingly
	fmt.Printf("Flag? Value: %v\n", *webFlag)
	if *webFlag {
		fmt.Println("Starting db monitor")
		go backend.StartFileCheckRoutine(&db, &tmpl)
	}

	// Listen on all network in  terfaces including localhost
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))

}

func combinedCardsHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	city1 := r.URL.Query().Get("city1")
	additionalCities := r.URL.Query()["city[]"]
	logicalOperators := r.URL.Query()["logical_operator[]"]
	maxPriceLinearStr := r.URL.Query().Get("maxPriceLinear")
	maxPriceLinear, err := strconv.ParseFloat(maxPriceLinearStr, 64)
	if err != nil {
		http.Error(w, "Invalid price parameter", http.StatusBadRequest)
		return
	}
	maxPrice := backend.MapLinearToExponential(maxPriceLinear, 50, 2500)

	// Determine sort option: 'nextCardsHandler' has a default, 'filterHandler' might vary
	sortOption := r.URL.Query().Get("sort")
	if sortOption == "" {
		sortOption = "low_price" // default for 'nextCardsHandler'
	}

	orderClause := determineOrderClause(sortOption)
	query := buildDynamicQuery(orderClause, city1, additionalCities, logicalOperators)
	params := append([]interface{}{city1, city1}, 1.0, 10.0, maxPrice) // Prepare parameters
	for _, city := range additionalCities {
		params = append(params, city)
	}

	rows, err := db.Query(query, params...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	flights, err := processFlightRows(rows)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	session.Values["city1"] = city1
	session.Save(r, w)

	data := buildFlightsData(city1, flights)
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
			&flight.BookingUrl,
			&flight.BookingPppn,
			&flight.FiveNightsFlights,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		// Use the image_1 URL from the database, or fallback to a placeholder if not available

		// Log the imageUrl for debugging
		log.Printf("Scanned image URL: '%s', Valid: %t", imageUrl.String, imageUrl.Valid)

		if imageUrl.Valid && len(imageUrl.String) > 5 {
			flight.RandomImageURL = imageUrl.String
			log.Printf("Using image URL from database: %s", flight.RandomImageURL)
		} else {
			flight.RandomImageURL = "/images/location-placeholder-image.png"
			log.Printf("Using default placeholder image URL: %s", flight.RandomImageURL)
		}

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
func buildFlightsData(city1 string, flights []Flight) FlightsData {
	var maxWpi, minFlightPrice, minHotelPrice, minFnafPrice sql.NullFloat64

	for _, flight := range flights {
		maxWpi = backend.UpdateMaxValue(maxWpi, flight.AvgWpi)
		minFlightPrice = backend.UpdateMinValue(minFlightPrice, flight.PriceCity1)
		minHotelPrice = backend.UpdateMinValue(minHotelPrice, flight.BookingPppn)
		minFnafPrice = backend.UpdateMinValue(minFnafPrice, flight.FiveNightsFlights)
	}

	return FlightsData{
		SelectedCity1: city1,
		Flights:       flights,
		MaxWpi:        maxWpi,
		MinFlight:     minFlightPrice,
		MinHotel:      minHotelPrice,
		MinFnaf:       minFnafPrice,
	}
}

// Unified Query Builder
func buildDynamicQuery(orderClause string, city1 string, additionalCities []string, logicalOperators []string) string {
	query := selectClause() +
		joinClause() +
		whereClause(city1, additionalCities, logicalOperators) +
		groupByClause() +
		havingClause() +
		orderClause

	// Print the query for debugging
	log.Printf("Generated Query: %s", query)
	return query
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
	default:
		return "ORDER BY fnf.price_fnaf ASC" // Default sorting by lowest FNAF price
	}
}

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
