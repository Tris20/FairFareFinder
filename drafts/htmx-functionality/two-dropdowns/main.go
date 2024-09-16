
package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"math"
  "fmt"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

type Flight struct {
	DestinationCityName string
	PriceCity1          sql.NullFloat64
	PriceCity2          sql.NullFloat64
	CombinedPrice       sql.NullFloat64
	UrlCity1            string
	UrlCity2            string
	AvgWpi              sql.NullFloat64
}

type FlightsData struct {
	SelectedCity1 string
	SelectedCity2 string
	Flights       []Flight
}

var (
	tmpl  *template.Template
	db    *sql.DB
	store *sessions.CookieStore = sessions.NewCookieStore([]byte("your-secret-key"))
)

func main() {
	var err error

	db, err = sql.Open("sqlite3", "./main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tmpl = template.Must(template.ParseFiles("index.html", "table.html"))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/filter", filterHandler)

  http.HandleFunc("/update-slider-price", updateSliderPriceHandler)
	
  log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Query to fetch distinct origin city names
	rows, err := db.Query("SELECT DISTINCT origin_city_name FROM flight")
	if err != nil {
		log.Printf("Error querying cities: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var cities []string
	for rows.Next() {
		var city string
		if err := rows.Scan(&city); err != nil {
			log.Printf("Error scanning city: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		cities = append(cities, city)
	}

	// Pass cities to template
	if err := tmpl.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"Cities": cities,
	}); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func filterHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	city1 := r.URL.Query().Get("city1")
	city2 := r.URL.Query().Get("city2")
	sortOption := r.URL.Query().Get("sort")
	minWpiStr := r.URL.Query().Get("wpi")           // Minimum WPI slider value
	maxPriceLinearStr := r.URL.Query().Get("maxPriceLinear") // Max price slider value (0-100 linear)

	// Convert string values to appropriate types
	minWpi, err := strconv.ParseFloat(minWpiStr, 64)
	if err != nil {
		log.Printf("Error parsing minWpi: %v", err)
		http.Error(w, "Invalid minWpi value", http.StatusBadRequest)
		return
	}

	// Convert the linear slider value (0-100) into an exponential scale (10-2500)
	maxPriceLinear, err := strconv.ParseFloat(maxPriceLinearStr, 64)
	if err != nil {
		log.Printf("Error parsing maxPriceLinear: %v", err)
		http.Error(w, "Invalid maxPrice value", http.StatusBadRequest)
		return
	}

	// Map the linear slider (0-100) to an exponential scale (10-2500)
	maxPrice := mapLinearToExponential(maxPriceLinear, 10, 2500)

	session.Values["city1"] = city1
	session.Values["city2"] = city2
	session.Save(r, w)

	orderClause := "ORDER BY combined_price ASC"
	switch sortOption {
	case "low_price":
		orderClause = "ORDER BY combined_price ASC"
	case "high_price":
		orderClause = "ORDER BY combined_price DESC"
	case "best_weather":
		orderClause = "ORDER BY avg_wpi DESC"
	case "worst_weather":
		orderClause = "ORDER BY avg_wpi ASC"
	}

	query := `
SELECT f1.destination_city_name, 
       MIN(f1.price_this_week) AS price_city1, 
       MIN(f2.price_this_week) AS price_city2,
       (MIN(f1.price_this_week) + MIN(f2.price_this_week)) AS combined_price,
       MIN(f1.skyscanner_url_this_week) AS url_city1, 
       MIN(f2.skyscanner_url_this_week) AS url_city2, 
       l.avg_wpi
FROM flight f1
JOIN flight f2 ON f1.destination_city_name = f2.destination_city_name
JOIN location l ON f1.destination_city_name = l.city
WHERE f1.origin_city_name = ? AND f2.origin_city_name = ? AND l.avg_wpi >= ?
GROUP BY f1.destination_city_name
HAVING combined_price <= ?
    ` + orderClause

	rows, err := db.Query(query, city1, city2, minWpi, maxPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var flights []Flight
	for rows.Next() {
		var flight Flight
		if err := rows.Scan(
			&flight.DestinationCityName,
			&flight.PriceCity1,
			&flight.PriceCity2,
			&flight.CombinedPrice,
			&flight.UrlCity1,
			&flight.UrlCity2,
			&flight.AvgWpi,
		); err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		flights = append(flights, flight)
	}

	data := FlightsData{
		SelectedCity1: city1,
		SelectedCity2: city2,
		Flights:       flights,
	}

	err = tmpl.ExecuteTemplate(w, "table.html", data)
	if err != nil {
		http.Error(w, "Error rendering results", http.StatusInternalServerError)
	}
}

// This function maps a linear slider (0-100) to an exponential range (10-2500)
func mapLinearToExponential(linearValue float64, minVal float64, maxVal float64) float64 {
    midVal := 1000.0
    percentage := linearValue / 100

    // First 70% of the slider covers 10 to 1000
    if percentage <= 0.7 {
        return minVal * math.Pow(midVal/minVal, percentage/0.7)
    }    else {

    // Last 30% covers 1000 to 2500
        newPercentage := (percentage - 0.7) / 0.3
        return midVal + (maxVal - midVal) * newPercentage
    }
}



func updateSliderPriceHandler(w http.ResponseWriter, r *http.Request) {
    // Get the linear slider value (0-100)
    maxPriceLinearStr := r.URL.Query().Get("maxPriceLinear")
    maxPriceLinear, err := strconv.ParseFloat(maxPriceLinearStr, 64)
    if err != nil {
        log.Printf("Error parsing maxPriceLinear: %v", err)
        http.Error(w, "Invalid maxPrice value", http.StatusBadRequest)
        return
    }

    // Map the linear value to the exponential price (10-2500)
    maxPrice := mapLinearToExponential(maxPriceLinear, 10, 2500)

    // Return the mapped value to be displayed next to the slider
    fmt.Fprintf(w, "%.2f", maxPrice)
}

