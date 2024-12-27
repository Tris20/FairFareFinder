package data_management

// database

import (
	"database/sql"
	"log"

	"github.com/Tris20/FairFareFinder/learning_utils_playground/config_handlers"
	db_manager "github.com/Tris20/FairFareFinder/learning_utils_playground/database_manager"
	_ "github.com/mattn/go-sqlite3"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/schollz/progressbar/v3"
)

// AirportInfo represents the basic information of an airport necessary for weather data fetching
type AirportInfo struct {
	City    string
	Country string
	IATA    string
}

// fetchAirports retrieves all airports with non-empty IATA codes from flights.db
func fetchAirports(db *sql.DB) ([]AirportInfo, error) {

	query := `SELECT a.city, a.country, a.iata
FROM airport a
JOIN city c ON LOWER(TRIM(a.city)) = LOWER(TRIM(c.city_ascii)) 
            AND LOWER(TRIM(a.country)) = LOWER(TRIM(c.iso2))  -- Using iso2 for country code
WHERE a.iata IS NOT NULL
AND a.iata != ''
AND a.city IS NOT NULL
AND a.country IS NOT NULL
AND c.include_tf = 1;
`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var airports []AirportInfo
	for rows.Next() {
		var ai AirportInfo
		if err := rows.Scan(&ai.City, &ai.Country, &ai.IATA); err != nil {
			return nil, err
		}
		airports = append(airports, ai)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return airports, nil
}

// initWeatherDB creates the weather database and table if it doesn't exist
func initWeatherDB(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Error opening weather.db: %v", err)
		return err
	}
	defer db.Close()

	err = db_manager.CreateTable(db, &db_manager.AllWeather{})
	if err != nil {
		log.Printf("Failed to create all_weather table: %v", err)
		return err
	}
	return nil
}

// weather.go

// WeatherDataFetch represents the structure of weather information to be stored in weather.db
type WeatherDataFetch struct {
	Date              string
	WeatherType       string
	Temperature       float64
	WeatherIconURL    string
	GoogleWeatherLink string
	WindSpeed         float64 // New field for wind speed
}

// fetchWeatherForCity fetches weather data for the specified city from OpenWeatherAPI
func fetchWeatherForCity(cityName string, countryCode string) ([]WeatherDataFetch, error) {
	// Placeholder for OpenWeatherAPI request. Assume you replace the following URL with the actual API request
	location_string := url.QueryEscape(fmt.Sprintf("%s, %s", cityName, countryCode))
	apiKey, err := config_handlers.LoadApiKey("../../../../ignore/secrets.yaml", "openweathermap.org")
	if err != nil {
		return nil, err
	}
	apiURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?q=%s&appid=%s&units=metric", location_string, apiKey)

	fmt.Printf("\napi url: %s \n", apiURL)

	/*
	     resp, err := http.Get(apiURL)
	   	if err != nil {
	   		return nil, err
	   	}
	   	defer resp.Body.Close()
	*/

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		return nil, fmt.Errorf("received non-200 status code from weather API: %d, response: %s", resp.StatusCode, bodyString)
	} else {
		fmt.Printf("\nWeather Data for %s %scollected successfully\n", cityName, countryCode)
	}

	var apiResp ApiResponse_Weather
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	var result []WeatherDataFetch // Assume proper JSON decoding based on OpenWeatherAPI response structure

	for _, item := range apiResp.List {
		// This example extracts weather data for each time entry in the list.
		// You might want to adjust this to extract daily averages or specific times of day.
		date := time.Unix(item.Dt, 0).Format("2006-01-02 15:04:05")
		temp := item.Main.Temp
		weatherType := "Clear" // Default to clear, adjust based on actual data
		if len(item.Weather) > 0 {
			weatherType = item.Weather[0].Main
		}
		iconURL := fmt.Sprintf("https://openweathermap.org/img/wn/%s.png", item.Weather[0].Icon)
		windSpeed := item.Wind.Speed // Extract wind speed

		googleWeatherURL := fmt.Sprintf("https://www.google.com/search?q=weather+%s", location_string)

		result = append(result, WeatherDataFetch{
			Date:              date,
			WeatherType:       weatherType,
			Temperature:       temp,
			WeatherIconURL:    iconURL,
			GoogleWeatherLink: googleWeatherURL,
			WindSpeed:         windSpeed, // Include wind speed
		})
	}

	return result, nil
}

func storeWeatherDataBatch(db *sql.DB, batch []WeatherDataBatch) error {
	// Set PRAGMA options for performance
	_, err := db.Exec("PRAGMA synchronous = OFF;")
	if err != nil {
		return fmt.Errorf("failed to set PRAGMA synchronous: %v", err)
	}
	_, err = db.Exec("PRAGMA journal_mode = MEMORY;")
	if err != nil {
		return fmt.Errorf("failed to set PRAGMA journal_mode: %v", err)
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback() // Rollback on error

	// Bulk insert statement
	query := `INSERT OR REPLACE INTO all_weather 
              (city_name, country_code, iata, date, weather_type, temperature, weather_icon_url, google_weather_link, wind_speed)
              VALUES `
	args := []interface{}{}

	for _, weatherDataBatch := range batch {
		for _, wd := range weatherDataBatch.WeatherInfo {
			query += "(?, ?, ?, ?, ?, ?, ?, ?, ?),"
			args = append(args,
				weatherDataBatch.Airport.City,
				weatherDataBatch.Airport.Country,
				weatherDataBatch.Airport.IATA,
				wd.Date,
				wd.WeatherType,
				wd.Temperature,
				wd.WeatherIconURL,
				wd.GoogleWeatherLink,
				wd.WindSpeed,
			)
		}
	}

	// Remove trailing comma and execute the bulk insert
	query = query[:len(query)-1]
	_, err = tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute bulk insert: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// main.go

type WeatherDataBatch struct {
	Airport     AirportInfo
	WeatherInfo []WeatherDataFetch
}

func FetchWeatherMain_UpdateWeatherDB(weatherDBPath, locationsDBPath string) error {

	var batch []WeatherDataBatch
	batchSize := 50
	flightsDB, err := sql.Open("sqlite3", locationsDBPath)

	if err != nil {
		log.Fatalf("Error opening locations.db: %v", err)
	}
	defer flightsDB.Close()

	// Initialize weather database
	initWeatherDB(weatherDBPath)

	// Open the database once
	db, err := sql.Open("sqlite3", weatherDBPath)
	if err != nil {
		log.Fatalf("Failed to open the database: %v", err)
	}
	defer db.Close()

	// Fetch airport info with non-empty IATA codes

	airports, err := fetchAirports(flightsDB)
	if err != nil {
		log.Fatalf("Error fetching airports: %v", err)
	}

	// Start of the rate limiting period
	//startTime := time.Now()

	// The maximum number of requests we can make per minute
	const maxRequestsPerMinute = 50
	// Calculate the interval at which we can make requests to not exceed the limit
	requestInterval := time.Minute / maxRequestsPerMinute

	// Create a new progress bar
	bar := progressbar.Default(int64(len(airports)))

	for _, airport := range airports {
		bar.Add(1)
		fmt.Printf("\ncity: %s  country: %s\n", airport.City, airport.Country)
		weatherInfo, err := fetchWeatherForCity(airport.City, airport.Country)
		if err != nil {
			log.Printf("Error fetching weather for %s: %v", airport.City, err)
			continue
		}

		// Add to batch
		batch = append(batch, WeatherDataBatch{
			Airport:     airport,
			WeatherInfo: weatherInfo,
		})

		if len(batch) >= batchSize {
			fmt.Println("Storing results...")
			if err := storeWeatherDataBatch(db, batch); err != nil {
				log.Printf("Error storing weather data for batch: %v", err)
			}
			batch = batch[:0] // Reset the batch
			fmt.Println("Batch stored")
		}

		// Rate-limiting logic (adjust as needed)
		time.Sleep(requestInterval)
	}

	// Insert any remaining batch
	if len(batch) > 0 {
		if err := storeWeatherDataBatch(db, batch); err != nil {
			log.Printf("Error storing weather data for final batch: %v", err)
		}
	}

	return nil
}
