
package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"math"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
  "fmt"
)

type Weather struct {
	Date             string          // Date for the weather forecast
	AvgDaytimeTemp   sql.NullFloat64 // Average daytime temperature
	WeatherIcon      string          // Weather icon URL
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

	tmpl = template.Must(template.ParseFiles("index.html", "table.html"))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/filter", filterHandler)
	http.HandleFunc("/update-slider-price", updateSliderPriceHandler)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))

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
	minWpiStr := r.URL.Query().Get("wpi")
	maxPriceLinearStr := r.URL.Query().Get("maxPriceLinear")

	// Convert string values to appropriate types
	minWpi, err := strconv.ParseFloat(minWpiStr, 64)
	if err != nil {
		log.Printf("Error parsing minWpi: %v", err)
		http.Error(w, "Invalid minWpi value", http.StatusBadRequest)
		return
	}

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

	// Updated query to join with the weather table for weather forecast
	query := `
SELECT f1.destination_city_name, 
       MIN(f1.price_this_week) AS price_city1, 
       MIN(f1.skyscanner_url_this_week) AS url_city1,
       w.date,
       w.avg_daytime_temp,
       w.weather_icon,
       w.avg_daytime_wpi
FROM flight f1
JOIN location l ON f1.destination_city_name = l.city
JOIN weather w ON w.city = f1.destination_city_name
WHERE f1.origin_city_name = ? AND l.avg_wpi >= ? AND w.date >= date('now')
GROUP BY f1.destination_city_name, w.date
HAVING price_city1 <= ?
    ` + orderClause

	rows, err := db.Query(query, city1, minWpi, maxPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var flights []Flight
	var currentFlight *Flight
	var lastCity string

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
			&weather.AvgDaytimeWpi,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if lastCity != flight.DestinationCityName {
			if currentFlight != nil {
				flights = append(flights, *currentFlight)
			}
			currentFlight = &Flight{
				DestinationCityName: flight.DestinationCityName,
				PriceCity1:          flight.PriceCity1,
				UrlCity1:            flight.UrlCity1,
				WeatherForecast:     []Weather{},
			}
			lastCity = flight.DestinationCityName
		}

		currentFlight.WeatherForecast = append(currentFlight.WeatherForecast, weather)
	}

	if currentFlight != nil {
		flights = append(flights, *currentFlight)
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
        return midVal + (maxVal - midVal) * newPercentage
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

    fmt.Fprintf(w, "%.2f", maxPrice)
}

