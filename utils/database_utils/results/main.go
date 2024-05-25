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
  fmt.Println("Inserting Weather Table")
	// Insert weather data into destination database
	err = InsertWeatherData(dbPath, weatherData)
	if err != nil {
		log.Fatalf("Failed to insert weather data: %v", err)
	}

	// Insert location data into destination database
  fmt.Println("Inserting Location Table")
  err = InsertLocationData(dbPath, weatherData)
	if err != nil {
		log.Fatalf("Failed to insert location data: %v", err)
	}


	log.Println("Weather data successfully transferred.")
}

