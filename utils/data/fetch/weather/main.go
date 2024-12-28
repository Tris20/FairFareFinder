package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

type WeatherDataBatch struct {
	Airport     AirportInfo
	WeatherInfo []WeatherData
}

func main() {

	var batch []WeatherDataBatch
	batchSize := 50
	flightsDB, err := sql.Open("sqlite3", "../../../../data/raw/locations/locations.db")

	if err != nil {
		log.Fatalf("Error opening locations.db: %v", err)
	}
	defer flightsDB.Close()

	// Initialize weather database
	weatherDBPath := "../../../../data/raw/weather/weather.db"
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

	count := 0
	for _, airport := range airports {
		count++
		if count > 2 {
			continue
		}
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
}
