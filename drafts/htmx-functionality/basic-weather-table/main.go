
package main

import (
	"database/sql"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
)

type Weather struct {
	Date             string          // Date for the weather forecast
	AvgDaytimeTemp   sql.NullFloat64 // Average daytime temperature
	WeatherIcon      string          // Weather icon URL
	GoogleUrl        string          // Google URL for the location
	AvgDaytimeWpi    sql.NullFloat64 // Weather Performance Index
}

type Flight struct {
	DestinationCityName string
	PriceCity1          sql.NullFloat64
	UrlCity1            string
	WeatherForecast     []Weather      // Slice of weather data for multiple days
}

type FlightsData struct {
	SelectedCity1 string
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

    // Parse templates
    tmpl = template.Must(template.ParseFiles("index.html", "table.html"))

    // Set up routes
    http.HandleFunc("/", indexHandler)                     // Homepage route
    http.HandleFunc("/filter", filterHandler)              // Route for filtering
    http.HandleFunc("/update-slider-price", updateSliderPriceHandler) // Route for slider price update
    http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css")))) // Serving CSS
    http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images")))) // Serving images

    // Privacy policy route
http.HandleFunc("/privacy-policy", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "privacy-policy.html")  // Make sure the path is correct
})

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
	sortOption := r.URL.Query().Get("sort")
//	minWpiStr := r.URL.Query().Get("wpi")
	maxPriceLinearStr := r.URL.Query().Get("maxPriceLinear")

//originCountry := r.URL.Query().Get("origin_country")
/*
	// Convert string values to appropriate types
	minWpi, err := strconv.ParseFloat(minWpiStr, 64)
	if err != nil {
		log.Printf("Error parsing minWpi: %v", err)
		http.Error(w, "Invalid minWpi value", http.StatusBadRequest)
		return
	}
*/
	maxPriceLinear, err := strconv.ParseFloat(maxPriceLinearStr, 64)
	if err != nil {
		log.Printf("Error parsing maxPriceLinear: %v", err)
		http.Error(w, "Invalid maxPrice value", http.StatusBadRequest)
		return
	}

	maxPrice := mapLinearToExponential(maxPriceLinear, 10, 2500)

	session.Values["city1"] = city1
	session.Save(r, w)

	orderClause := "ORDER BY price_city1 ASC"
	switch sortOption {
	case "low_price":
		orderClause = "ORDER BY price_city1 ASC"
	case "high_price":
		orderClause = "ORDER BY price_city1 DESC"
	case "best_weather":
		orderClause = "ORDER BY avg_wpi DESC"
	case "worst_weather":
		orderClause = "ORDER BY avg_wpi ASC"
	}

	// Calculate the lower and upper bounds for WPI
//	lowerWpi := math.Max(minWpi-2.5, 1.0)   // Lower bound constrained to 1.0
//	upperWpi := math.Min(minWpi+2.5, 10.0)  // Upper bound constrained to 10.0
// Disabling slider for now
lowerWpi := 1.0 
upperWpi := 10.0


	// Updated query to join with the weather table for weather forecast

query := `
SELECT f1.destination_city_name, 
       MIN(f1.price_this_week) AS price_city1, 
       MIN(f1.skyscanner_url_this_week) AS url_city1,
       w.date,
       w.avg_daytime_temp,
       w.weather_icon,
       w.google_url,
       w.avg_daytime_wpi
FROM flight f1
JOIN location l ON f1.destination_city_name = l.city AND f1.destination_country = l.country
JOIN weather w ON w.city = f1.destination_city_name AND w.country = f1.destination_country
WHERE f1.origin_city_name = ? 
AND l.avg_wpi BETWEEN ? AND ? 
AND w.date >= date('now')
GROUP BY f1.destination_city_name, w.date, f1.destination_country /* Add country to GROUP BY */
HAVING price_city1 <= ?
    ` + orderClause

rows, err := db.Query(query, city1, lowerWpi, upperWpi, maxPrice)
if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}
defer rows.Close()

	var flights []Flight

	// Loop through the rows and construct flights and weather forecasts
	for rows.Next() {
		var flight Flight
		var weather Weather

		err := rows.Scan(
			&flight.DestinationCityName,
			&flight.PriceCity1,
			&flight.UrlCity1,
			&weather.Date,
			&weather.AvgDaytimeTemp,
			&weather.WeatherIcon,
			&weather.GoogleUrl,
			&weather.AvgDaytimeWpi,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Check if the destination already exists in flights
		found := false
		for i := range flights {
			if flights[i].DestinationCityName == flight.DestinationCityName {
				// Append the weather forecast to the existing flight entry
				flights[i].WeatherForecast = append(flights[i].WeatherForecast, weather)
				found = true
				break
			}
		}

		if !found {
			// Add a new flight entry if this destination wasn't found in the list
			flight.WeatherForecast = []Weather{weather}
			flights = append(flights, flight)
		}
	}

	data := FlightsData{
		SelectedCity1: city1,
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
	} else {
		// Last 30% covers 1000 to 2500
		newPercentage := (percentage - 0.7) / 0.3
		return midVal + (maxVal-midVal)*newPercentage
	}
}

func updateSliderPriceHandler(w http.ResponseWriter, r *http.Request) {
	maxPriceLinearStr := r.URL.Query().Get("maxPriceLinear")
	maxPriceLinear, err := strconv.ParseFloat(maxPriceLinearStr, 64)
	if err != nil {
		log.Printf("Error parsing maxPriceLinear: %v", err)
		http.Error(w, "Invalid maxPrice value", http.StatusBadRequest)
		return
	}

	maxPrice := mapLinearToExponential(maxPriceLinear, 10, 2500)

	fmt.Fprintf(w, "€%.2f", maxPrice)
}

