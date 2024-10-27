
package main

import (
	"database/sql"
	"flag"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"fmt"

	"github.com/Tris20/FairFareFinder/src/backend"
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

	// Parse the "web" flag
	webFlag := flag.Bool("web", false, "Pass this flag to enable the web server with file check routine")
	flag.Parse() // Parse command-line flags

	var err error

	db, err = sql.Open("sqlite3", "./data/compiled/main.db")

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Parse templates
	tmpl = template.Must(template.ParseFiles("./src/frontend/html/index.html", "./src/frontend/html/table.html"))

	backend.Init(db, tmpl)

	// Set up routes
	http.HandleFunc("/", backend.IndexHandler)
	http.HandleFunc("/filter", filterHandler) // Route for filtering
	http.HandleFunc("/next-cards", nextCardsHandler) // New route for loading more cards
	http.HandleFunc("/update-slider-price", backend.UpdateSliderPriceHandler) // Route for slider price update
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./src/frontend/css/"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./src/frontend/images"))))

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

	// Listen on all network interfaces including localhost
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))

}

func filterHandler(w http.ResponseWriter, r *http.Request) {
	// Same as existing filterHandler logic
	session, _ := store.Get(r, "session")
	city1, sortOption, maxPriceLinear, err := parseFilterRequest(r)
	if err != nil {
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		return
	}

	maxPrice := backend.MapLinearToExponential(maxPriceLinear, 100, 2500)
	session.Values["city1"] = city1
	session.Save(r, w)

	orderClause := determineOrderClause(sortOption)
	query := buildFilterQuery(orderClause)

	rows, err := db.Query(query, city1, city1, 1.0, 10.0, maxPrice)
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

	data := buildFlightsData(city1, flights)
	err = tmpl.ExecuteTemplate(w, "table.html", data)
	if err != nil {
		http.Error(w, "Error rendering results", http.StatusInternalServerError)
	}
}


func nextCardsHandler(w http.ResponseWriter, r *http.Request) {
    // Get pagination parameters (like offset, limit) from the query params
    pageStr := r.URL.Query().Get("page")
    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 {
        page = 1 // Default to first page if invalid
    }

    offset := (page - 1) * 10 // Assuming 10 results per page
    limit := 10

    // Fetch the city and maximum price parameters from the query string
    city1 := r.URL.Query().Get("city1")
    maxPriceLinearStr := r.URL.Query().Get("maxPriceLinear")
    maxPriceLinear, err := strconv.ParseFloat(maxPriceLinearStr, 64)
    if err != nil {
        http.Error(w, "Invalid price parameter", http.StatusBadRequest)
        return
    }

    maxPrice := backend.MapLinearToExponential(maxPriceLinear, 100, 2500)

    // Updated query to ensure origin city matches properly
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
    ORDER BY fnf.price_fnaf ASC
    LIMIT ? OFFSET ?`

    // Execute the query with the appropriate parameters
    rows, err := db.Query(query, city1, city1, 1.0, 10.0, maxPrice, limit, offset)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    flights, err := processFlightRows(rows)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Append more flights as new cards to the carousel
    err = tmpl.ExecuteTemplate(w, "table.html", flights)
    if err != nil {
        http.Error(w, "Error rendering results", http.StatusInternalServerError)
    }
}

// The rest of the code remains the same (helper functions, etc.)


// Helper function to parse request parameters
func parseFilterRequest(r *http.Request) (string, string, float64, error) {
	city1 := r.URL.Query().Get("city1")
	sortOption := r.URL.Query().Get("sort")
	maxPriceLinearStr := r.URL.Query().Get("maxPriceLinear")

	maxPriceLinear, err := strconv.ParseFloat(maxPriceLinearStr, 64)
	if err != nil {
		log.Printf("Error parsing maxPriceLinear: %v", err)
		return "", "", 0, err
	}
	return city1, sortOption, maxPriceLinear, nil
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

// Helper function to build the query string
func buildFilterQuery(orderClause string) string {
	return `
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
HAVING fnf.price_fnaf <= ? ` + orderClause
}

// Helper function to process rows into flight and weather data
func processFlightRows(rows *sql.Rows) ([]Flight, error) {
	var flights []Flight
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
			&flight.AvgWpi,
			&flight.BookingUrl,
			&flight.BookingPppn,
			&flight.FiveNightsFlights,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
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
		maxWpi = updateMaxValue(maxWpi, flight.AvgWpi)
		minFlightPrice = updateMinValue(minFlightPrice, flight.PriceCity1)
		minHotelPrice = updateMinValue(minHotelPrice, flight.BookingPppn)
		minFnafPrice = updateMinValue(minFnafPrice, flight.FiveNightsFlights)
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

// Helper function to update max value
func updateMaxValue(currentMax, newValue sql.NullFloat64) sql.NullFloat64 {
	if !currentMax.Valid || (newValue.Valid && newValue.Float64 > currentMax.Float64) {
		return newValue
	}
	return currentMax
}

// Helper function to update min value
func updateMinValue(currentMin, newValue sql.NullFloat64) sql.NullFloat64 {
	// HOTFIX Check if newValue is valid and greater than or equal to 0.1
  // This ensures we don't include flight prices which are zero because no price was found  
	if newValue.Valid && newValue.Float64 >= 0.1 {
		// Update currentMin if it's not valid or if newValue is smaller
		if !currentMin.Valid || newValue.Float64 < currentMin.Float64 {
			return newValue
		}
	}
	// Return currentMin if none of the above conditions are met
	return currentMin
}

