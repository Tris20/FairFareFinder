
package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)



type Flight struct {
	DestinationCityName string
	PriceCity1          sql.NullFloat64
	PriceCity2          sql.NullFloat64
	CombinedPrice       sql.NullFloat64
	AvgWpi              sql.NullFloat64
}



type FlightsData struct {
	SelectedCity1 string
	SelectedCity2 string
	Flights       []Flight
}

var tmpl *template.Template
var db *sql.DB

// Global variables to store selected cities
var selectedCity1 string
var selectedCity2 string

func main() {
	var err error

	// Connect to SQLite database (main.db)
	db, err = sql.Open("sqlite3", "./main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Parse both index.html and table.html templates
	tmpl = template.Must(template.ParseFiles("index.html", "table.html"))

	// Handlers
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/filter", filterHandler)

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Get distinct origin city names for the dropdown
	rows, err := db.Query("SELECT DISTINCT origin_city_name FROM flight")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var cities []string
	for rows.Next() {
		var city string
		if err := rows.Scan(&city); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cities = append(cities, city)
	}

	// Render the main page with the dropdown and empty table section
	err = tmpl.ExecuteTemplate(w, "index.html", cities)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}





func filterHandler(w http.ResponseWriter, r *http.Request) {
	city1 := r.URL.Query().Get("city1")
	city2 := r.URL.Query().Get("city2")
	sortOption := r.URL.Query().Get("sort")

	// Update global variables based on what was selected
	if city1 != "" {
		selectedCity1 = city1
	}
	if city2 != "" {
		selectedCity2 = city2
	}

	// Default sorting: lowest to highest combined price
	orderClause := "ORDER BY combined_price ASC"

	// Adjust the SQL sorting based on the selected sort option
	switch sortOption {
	case "low_price":
		orderClause = "ORDER BY combined_price ASC"
	case "high_price":
		orderClause = "ORDER BY combined_price DESC"
	case "best_weather":
		orderClause = "ORDER BY l.avg_wpi DESC"
	case "worst_weather":
		orderClause = "ORDER BY l.avg_wpi ASC"
	}

	var query string
	var rows *sql.Rows
	var err error

	if selectedCity1 != "" && selectedCity2 != "" {
		// Query to get destinations with the lowest price from both cities, avg_wpi from the location table, and combined price
		query = `
			SELECT f1.destination_city_name, MIN(f1.price_this_week), MIN(f2.price_this_week),
			(MIN(f1.price_this_week) + MIN(f2.price_this_week)) AS combined_price, l.avg_wpi
			FROM flight f1
			INNER JOIN flight f2 ON f1.destination_city_name = f2.destination_city_name
			INNER JOIN location l ON f1.destination_city_name = l.city
			WHERE f1.origin_city_name = ? AND f2.origin_city_name = ?
			GROUP BY f1.destination_city_name
			` + orderClause
		rows, err = db.Query(query, selectedCity1, selectedCity2)
		fmt.Println("Query for both cities:", selectedCity1, selectedCity2)
	} else if selectedCity1 != "" {
		// If only city1 is selected, show the lowest price for flights from that city and avg_wpi from location
		query = `
			SELECT f.destination_city_name, MIN(f.price_this_week), NULL, MIN(f.price_this_week), l.avg_wpi
			FROM flight f
			INNER JOIN location l ON f.destination_city_name = l.city
			WHERE f.origin_city_name = ?
			GROUP BY f.destination_city_name
			` + orderClause
		rows, err = db.Query(query, selectedCity1)
		fmt.Println("Query for single city:", selectedCity1)
	} else if selectedCity2 != "" {
		// If only city2 is selected, show the lowest price for flights from that city and avg_wpi from location
		query = `
			SELECT f.destination_city_name, NULL, MIN(f.price_this_week), MIN(f.price_this_week), l.avg_wpi
			FROM flight f
			INNER JOIN location l ON f.destination_city_name = l.city
			WHERE f.origin_city_name = ?
			GROUP BY f.destination_city_name
			` + orderClause
		rows, err = db.Query(query, selectedCity2)
		fmt.Println("Query for single city:", selectedCity2)
	}

	if err != nil {
		fmt.Println("Error executing query:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var flights []Flight
	for rows.Next() {
		var flight Flight
		err := rows.Scan(&flight.DestinationCityName, &flight.PriceCity1, &flight.PriceCity2, &flight.CombinedPrice, &flight.AvgWpi)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		flights = append(flights, flight)
	}

	if len(flights) == 0 {
		fmt.Println("No flights found for cities:", selectedCity1, selectedCity2)
	} else {
		fmt.Println("Flights found:", flights)
	}

	// Prepare the data to be sent to the template
	data := FlightsData{
		SelectedCity1: selectedCity1,
		SelectedCity2: selectedCity2,
		Flights:       flights,
	}

	// Render the partial table template with the filtered flight results
	err = tmpl.ExecuteTemplate(w, "table.html", data)
	if err != nil {
		fmt.Println("Error rendering template:", err)
		http.Error(w, "Error rendering results", http.StatusInternalServerError)
	}
}

