package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/Tris20/FairFareFinder/src/backend"
)

type Weather struct {
	Date           string          // Date for the weather forecast
	AvgDaytimeTemp sql.NullFloat64 // Average daytime temperature
	WeatherIcon    string          // Weather icon URL
	GoogleUrl      string          // Google URL for the location
	AvgDaytimeWpi  sql.NullFloat64 // Weather Performance Index
}

type Flight struct {
	DestinationCityName string
	PriceCity1          sql.NullFloat64
	UrlCity1            string
	WeatherForecast     []Weather
	AvgWpi              sql.NullFloat64 // Add this field for avg_wpi from the location table
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

	// Parse the "web" flag
	webFlag := flag.Bool("web", false, "Pass this flag to enable the web server with file check routine")
	flag.Parse() // Parse command-line flags

	var err error

	//	db, err = sql.Open("sqlite3", "./main.db")
	db, err = sql.Open("sqlite3", "./data/compiled/main.db")

	if err != nil {

		log.Fatal(err)
	}
	defer db.Close()

	// Parse templates
	tmpl = template.Must(template.ParseFiles("./src/frontend/html/index.html", "./src/frontend/html/table.html"))

	backend.Init(db, tmpl)

	// Set up routes
	http.HandleFunc("/", backend.IndexHandler)                                                                // Homepage route
	http.HandleFunc("/filter", filterHandler)                                                                 // Route for filtering
	http.HandleFunc("/update-slider-price", updateSliderPriceHandler)                                         // Route for slider price update
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./src/frontend/css/"))))         // Serving CSS
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./src/frontend/images")))) // Serving images

	// Privacy policy route
	http.HandleFunc("/privacy-policy", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./src/frontend/html/privacy-policy.html") // Make sure the path is correct
	})

	// On web server, every 2 hours, check for a new database deilvery, and swap dbs accordingly
	if *webFlag {
		// Start the file checking routine
		go backend.StartFileCheckRoutine(db)
	}

	// liston on all newtowrk interfaces including localhost
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))

}

func filterHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	city1 := r.URL.Query().Get("city1")
	sortOption := r.URL.Query().Get("sort")
	//	minWpiStr := r.URL.Query().Get("wpi")
	maxPriceLinearStr := r.URL.Query().Get("maxPriceLinear")

	maxPriceLinear, err := strconv.ParseFloat(maxPriceLinearStr, 64)
	if err != nil {
		log.Printf("Error parsing maxPriceLinear: %v", err)
		http.Error(w, "Invalid maxPrice value", http.StatusBadRequest)
		return
	}

	maxPrice := mapLinearToExponential(maxPriceLinear, 100, 2500)

	session.Values["city1"] = city1
	session.Save(r, w)

	orderClause := "ORDER BY fnf.price_fnaf ASC" // Default to sorting by FNAF price in ascending order
	switch sortOption {
	case "low_price":
		orderClause = "ORDER BY fnf.price_fnaf ASC" // Sort by lowest FNAF price
	case "high_price":
		orderClause = "ORDER BY fnf.price_fnaf DESC" // Sort by highest FNAF price
	case "best_weather":
		orderClause = "ORDER BY avg_wpi DESC" // Sort by best weather (highest WPI)
	case "worst_weather":
		orderClause = "ORDER BY avg_wpi ASC" // Sort by worst weather (lowest WPI)
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
       l.avg_wpi,  
       a.booking_url,
       a.booking_pppn,
       fnf.price_fnaf 
FROM flight f1
JOIN location l ON f1.destination_city_name = l.city AND f1.destination_country = l.country
JOIN weather w ON w.city = f1.destination_city_name AND w.country = f1.destination_country
LEFT JOIN accommodation a ON a.city = f1.destination_city_name AND a.country = f1.destination_country
LEFT JOIN five_nights_and_flights fnf ON fnf.destination_city = f1.destination_city_name AND fnf.origin_city = ? 
WHERE f1.origin_city_name = ? 
AND l.avg_wpi BETWEEN ? AND ? 
AND w.date >= date('now')
GROUP BY f1.destination_city_name, w.date, f1.destination_country, l.avg_wpi 
HAVING fnf.price_fnaf <= ?

` + orderClause

	rows, err := db.Query(query, city1, city1, lowerWpi, upperWpi, maxPrice)

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
			&flight.AvgWpi, // Add this to scan avg_wpi from the location table
			&flight.BookingUrl,
			&flight.BookingPppn,
			&flight.FiveNightsFlights, // Scan price_fnaf
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

	// Initialize variables to track the highest and lowest values
	var maxWpi sql.NullFloat64
	var minFlightPrice, minHotelPrice, minFnafPrice sql.NullFloat64

	for _, flight := range flights {
		if !maxWpi.Valid || (flight.AvgWpi.Valid && flight.AvgWpi.Float64 > maxWpi.Float64) {
			maxWpi = flight.AvgWpi // Set the highest WPI found
		}
		if !minFlightPrice.Valid || (flight.PriceCity1.Valid && flight.PriceCity1.Float64 < minFlightPrice.Float64) {
			minFlightPrice = flight.PriceCity1
		}
		if !minHotelPrice.Valid || (flight.BookingPppn.Valid && flight.BookingPppn.Float64 < minHotelPrice.Float64) {
			minHotelPrice = flight.BookingPppn
		}
		if !minFnafPrice.Valid || (flight.FiveNightsFlights.Valid && flight.FiveNightsFlights.Float64 < minFnafPrice.Float64) {
			minFnafPrice = flight.FiveNightsFlights
		}
	}

	// Pass these values to the template
	data := FlightsData{
		SelectedCity1: city1,
		Flights:       flights,
		MaxWpi:        maxWpi,         // Add highest WPI
		MinFlight:     minFlightPrice, // Add lowest flight price
		MinHotel:      minHotelPrice,  // Add lowest avg hotel price
		MinFnaf:       minFnafPrice,   // Add lowest FNAF price
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

	maxPrice := mapLinearToExponential(maxPriceLinear, 100, 2500)

	fmt.Fprintf(w, "€%.2f", maxPrice)
}
