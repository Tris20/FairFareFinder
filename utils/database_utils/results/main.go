package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// Create /out directory if it does not exist
	outputDir := "./out"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.Mkdir(outputDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create directory %s: %v", outputDir, err)
		}
	}

	// Database file paths
	dbPath := filepath.Join(outputDir, "results.db")
	sourceDBPath := "../../../data/longterm_db/weather.db"

	// Backup existing database if it exists
	backupDatabase(dbPath, outputDir)

	// Initialize the new database and create tables
	initializeDatabase(dbPath)

	// Fetch weather data from source database
	weatherData, err := FetchWeatherData(sourceDBPath)
	if err != nil {
		log.Fatalf("Failed to fetch weather data: %v", err)
	}


	// Insert weather data into destination database
fmt.Println("Populating weather_detailed table")
	err = InsertWeatherData("weather_detailed", dbPath, weatherData)
	if err != nil {
		log.Fatalf("Failed to insert weather data: %v", err)
	}

	// Prepare unique locations
	uniqueLocations, err := PrepareLocationData(weatherData)
	if err != nil {
		log.Fatalf("Failed to prepare unique locations: %v", err)
	}

  // Collect daily average weather records
	dailyAverageWeatherRecords, err := CollectDailyAverageWeather(weatherData)
	if err != nil {
		log.Fatalf("Failed to collect daily average weather: %v", err)
	}

  // Create and Populate the Daily Average Table
fmt.Println("Populating weather_daily_average table")
	err = InsertWeatherData("weather_daily_average", dbPath, dailyAverageWeatherRecords)
	if err != nil {
		log.Fatalf("Failed to insert weather data: %v", err)
	}



fmt.Println("Inserting Locations")
	// Insert location data into destination database
	err = InsertLocationData(dbPath, uniqueLocations)
	if err != nil {
		log.Fatalf("Failed to insert location data: %v", err)
	}

	log.Println("Weather and location data successfully transferred.")
}

