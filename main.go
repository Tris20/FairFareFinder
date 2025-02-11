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

	// Third-Party Packages
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/natefinch/lumberjack.v2"

	// Local Packages
	"github.com/Tris20/FairFareFinder/src/backend"
	"github.com/Tris20/FairFareFinder/src/backend/config"
)

// Global variables: template, database, session store
var (
	tmpl  *template.Template
	db    *sql.DB
	store *sessions.CookieStore = sessions.NewCookieStore([]byte("your-secret-key"))
)

func main() {
	config.SetMutePrints(false)
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

	backend.SetupRoutes(store, db, tmpl)
	//load filterRequeasthandler separately because it still lives in main
	http.HandleFunc("/filter", filterRequestHandler)

	return cleanup
}

func StartServer() {
	// Listen on all network interfaces including localhost
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

func filterRequestHandler(w http.ResponseWriter, r *http.Request) {
	//  Session Management
	session, err := backend.GetUserSession(store, r)
	if err != nil {
		backend.HandleHTTPError(w, "Session retrieval error", http.StatusInternalServerError)
		return
	}

	//  Input Extraction and Validation
	input, err := backend.ParseAndValidateFilterInputs(r)
	if err != nil {
		backend.HandleHTTPError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Execute Main Query to Populate Destination Cards
	flights, err := backend.ExecuteMainQuery(input)
	if err != nil {
		backend.HandleHTTPError(w, "Error executing main query", http.StatusInternalServerError)
		return
	}

	//  Execute Second Query to Populate Accommodation Price Slider Histogram
	allAccomPrices, err := backend.ExecuteAccommodationPricesHistogramQuery(input)
	if err != nil {
		backend.HandleHTTPError(w, "Error executing all prices query", http.StatusInternalServerError)
		return
	}
	log.Printf("All accommodation prices (no user limit): %v", allAccomPrices)

	//  Execute Second Query to Populate Flight Price Slider Histogram

	// Execute the query which returns an array of Flight structs.
	flightsPriceHistogramData, err := backend.ExecuteFlightPricesHistogramQuery(input)
	if err != nil {
		backend.HandleHTTPError(w, "Error executing all prices query", http.StatusInternalServerError)
		return
	}

	var allFlightPrices [][]float64

	for _, city := range input.Cities {
		var flightPricesForCity []float64

		for _, flight := range flightsPriceHistogramData {
			if flight.UrlCity1 == city {
				if flight.PriceCity1.Valid {
					flightPricesForCity = append(flightPricesForCity, flight.PriceCity1.Float64)
				} else {
					// Optionally append 0.0 or skip if invalid.
					flightPricesForCity = append(flightPricesForCity, 0.0)
				}
			}
		}
		log.Printf("Filtered Flight Prices for %s: %v", city, flightPricesForCity)
		allFlightPrices = append(allFlightPrices, flightPricesForCity)
	}

	//  Prepare Data for the Template
	data := backend.BuildTemplateData(input.Cities, flights, allAccomPrices, allFlightPrices)
	fmt.Printf("AllFlightPrices: %+v\n", data.AllFlightPrices)
	// Save Session and Render the Response
	if err := session.Save(r, w); err != nil {
		backend.HandleHTTPError(w, "Session save error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "table.html", data); err != nil {
		backend.HandleHTTPError(w, "Error rendering results", http.StatusInternalServerError)
		return
	}
}
