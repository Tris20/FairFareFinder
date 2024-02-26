package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func main() {
	dbFilepath := "flights.db"
	inputDir := "input"

	// Initialize the database
	db := InitializeDB(dbFilepath)
	defer db.Close()

	// Get a list of JSON files in the input directory
	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		log.Fatalf("Failed to read input directory: %v", err)
	}

	// Process each file
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(inputDir, file.Name())
			log.Printf("Processing file: %s", filePath)

			// Check if the file is empty
			info, err := os.Stat(filePath)
			if err != nil {
				log.Printf("Error getting file info: %v", err)
				continue
			}
			if info.Size() == 0 {
				log.Printf("Skipping empty file: %s", filePath)
				continue
			}

			// Parse JSON file

			apiResponse, err := ParseJSON(filePath)
			if err != nil {
				log.Printf("Error parsing JSON data: %v", err)
				continue
			}

			// Insert data into the database
			for _, departure := range apiResponse.Departures {
				InsertFlightData(db, "Departure", departure.Departure.Airport.IATA, departure.Departure.ScheduledTime.Local, departure.Arrival.Airport.IATA, departure.Arrival.ScheduledTime.Local)
			}
			// Assume similar logic for arrivals if applicable
		}
	}

	log.Println("All flight data inserted successfully.")
}
