
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
	OriginCityName      string
	OriginCountry       string
	DestinationCityName string
	DestinationCountry  string
	PriceThisWeek       float64
	DurationInMinutes   sql.NullFloat64
}

var tmpl *template.Template
var db *sql.DB

// Store the selected city values in memory (per session or globally)
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
	// Read the cities from the query parameters
	newCity1 := r.URL.Query().Get("city1")
	newCity2 := r.URL.Query().Get("city2")

	// Update the selected cities in memory
	if newCity1 != "" {
		selectedCity1 = newCity1
	}
	if newCity2 != "" {
		selectedCity2 = newCity2
	}

	// Debugging: Log the current selected cities
	fmt.Println("Selected cities:", selectedCity1, selectedCity2)

	var query string
	var rows *sql.Rows
	var err error

	// Handle different cases of selected cities
	switch {
	case selectedCity1 != "" && selectedCity2 != "":
		// Query for flights from both cities where destination is common
		query = `
			SELECT f1.destination_city_name, f1.destination_country, f1.price_this_week, f1.duration_in_minutes
			FROM flight f1
			INNER JOIN flight f2 ON f1.destination_city_name = f2.destination_city_name
			WHERE f1.origin_city_name = ? AND f2.origin_city_name = ?
			GROUP BY f1.destination_city_name`
		rows, err = db.Query(query, selectedCity1, selectedCity2)
	case selectedCity1 != "":
		// Query for flights from city1 only
		query = `
			SELECT destination_city_name, destination_country, price_this_week, duration_in_minutes
			FROM flight
			WHERE origin_city_name = ?`
		rows, err = db.Query(query, selectedCity1)
	case selectedCity2 != "":
		// Query for flights from city2 only
		query = `
			SELECT destination_city_name, destination_country, price_this_week, duration_in_minutes
			FROM flight
			WHERE origin_city_name = ?`
		rows, err = db.Query(query, selectedCity2)
	default:
		http.Error(w, "No city selected", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var flights []Flight
	for rows.Next() {
		var flight Flight
		err := rows.Scan(&flight.DestinationCityName, &flight.DestinationCountry, &flight.PriceThisWeek, &flight.DurationInMinutes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		flights = append(flights, flight)
	}

	// Render the table with the filtered flight results
	err = tmpl.ExecuteTemplate(w, "table.html", flights)
	if err != nil {
		http.Error(w, "Error rendering results", http.StatusInternalServerError)
	}
}

