
package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/schollz/progressbar/v3"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open connection to flights.db
	flightsDB, err := sql.Open("sqlite3", "../../../../data/longterm_db/flights.db")
	if err != nil {
		log.Fatalf("Error opening flights.db: %v", err)
	}
	defer flightsDB.Close()

	// Initialize weather database
	weatherDBPath := "../../../../data/longterm_db/weather.db"
	initWeatherDB(weatherDBPath)

	// Fetch airport info with non-empty IATA codes
	airports, err := fetchAirports(flightsDB)
	if err != nil {
		log.Fatalf("Error fetching airports: %v", err)
	}

	// Start of the rate limiting period
	startTime := time.Now()

	// The maximum number of requests we can make per minute
	const maxRequestsPerMinute = 500
	// Calculate the interval at which we can make requests to not exceed the limit
	requestInterval := time.Minute / maxRequestsPerMinute

	// Create a new progress bar
	bar := progressbar.Default(int64(len(airports)))

	for i, airport := range airports {
		// Update the progress bar
		bar.Add(1)

		if i > 0 && i%maxRequestsPerMinute == 0 {
			// Calculate how much time has passed since the start of the rate-limiting period
			elapsed := time.Since(startTime)
			// If we've made 50 requests before a minute has passed, wait for the remainder of the minute
			if elapsed < time.Minute {
				time.Sleep(time.Minute - elapsed)
			}
			// Reset the start time for the next batch of requests
			startTime = time.Now()
		}

		weatherInfo, err := fetchWeatherForCity(airport.City, airport.Country)
		if err != nil {
			log.Printf("Error fetching weather for %s: %v", airport.City, err)
			continue
		}

		if err := storeWeatherData(weatherDBPath, airport, weatherInfo); err != nil {
			log.Printf("Error storing weather data for %s: %v", airport.City, err)
		}

		// Wait for the calculated request interval before making the next request
		time.Sleep(requestInterval)
	}
}

