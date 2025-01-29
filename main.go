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
	"strings"

	// Third-Party Packages
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/natefinch/lumberjack.v2"

	// Local Packages
	"github.com/Tris20/FairFareFinder/src/backend"
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
	flights, err := executeMainQuery(input)
	if err != nil {
		backend.HandleHTTPError(w, "Error executing main query", http.StatusInternalServerError)
		return
	}

	//  Execute Second Query to Populate Accommodation Price Slider Histogram
	allAccomPrices, err := executeAccommodationPricesHistogramQuery(input)
	if err != nil {
		backend.HandleHTTPError(w, "Error executing all prices query", http.StatusInternalServerError)
		return
	}

	log.Printf("All accommodation prices (no user limit): %v", allAccomPrices)

	//  Prepare Data for the Template
	data := backend.BuildTemplateData(input.Cities, flights, allAccomPrices)

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

/*
#
#
#
*/

func executeMainQuery(input *backend.FilterInput) ([]model.Flight, error) {
	query, args := buildQuery(input.LogicalExpression, input.MaxAccommodationPrice, input.Cities, input.OrderClause)

	if !mutePrints {
		fmt.Println("Generated SQL Query (MAIN):")
		fmt.Println(query)
		fmt.Println("Arguments:", args)
	}

	fullQuery := backend.InterpolateQuery(query, args)
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

	flights, err := backend.ProcessFlightRows(rows)
	if err != nil {
		log.Printf("Error processing flight rows: %v", err)
		return nil, err
	}

	return flights, nil
}

func executeAccommodationPricesHistogramQuery(input *backend.FilterInput) ([]float64, error) {
	// Build query with fixed maxAccommodationPrice = 550.0
	allPricesQuery, allPricesArgs := buildQuery(input.LogicalExpression, 550.0, input.Cities, input.OrderClause)

	if !mutePrints {
		fmt.Println("Generated SQL Query (ALL PRICES):")
		fmt.Println(allPricesQuery)
		fmt.Println("Arguments:", allPricesArgs)
	}

	fullAllPricesQuery := backend.InterpolateQuery(allPricesQuery, allPricesArgs)
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

	flightsAll, err := backend.ProcessFlightRows(rows)
	if err != nil {
		log.Printf("Error processing flight rows for all accommodation prices: %v", err)
		return nil, err
	}

	var prices []float64
	for _, f := range flightsAll {
		if f.BookingPppn.Valid {
			prices = append(prices, f.BookingPppn.Float64)
		}
	}
	return prices, nil
}

/*
#
#
#
*/

// Unified Query Builder

func buildQuery(expr backend.Expression, maxAccommodationPrice float64, originCities []string, orderClause string) (string, []interface{}) {
	var queryBuilder strings.Builder
	var args []interface{}
	// Begin the query with the DestinationSet CTE
	queryBuilder.WriteString("WITH DestinationSet AS (\n")

	// Build the subquery based on the logical expression
	subquery, subqueryArgs := buildSubquery(expr)
	queryBuilder.WriteString(subquery)
	queryBuilder.WriteString("\n)")
	args = append(args, subqueryArgs...)

	// This is where the core part of the sql query comes from
	queryBuilder.WriteString(backend.BaseQuery)

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

func buildSubquery(expr backend.Expression) (string, []interface{}) {
	switch e := expr.(type) {
	case *backend.CityCondition:
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
	case *backend.LogicalExpression:
		// Build the left and right subqueries
		leftSubquery, leftArgs := buildSubquery(e.Left)
		rightSubquery, rightArgs := buildSubquery(e.Right)
		var operator string
		if e.Operator == backend.AndOperator {
			operator = "INTERSECT"
		} else if e.Operator == backend.OrOperator {
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
