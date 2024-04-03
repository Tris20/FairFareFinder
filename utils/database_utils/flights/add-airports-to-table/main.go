package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

// CorrectCityMap provides a mapping for airports where the city information in the CSV might not be accurate.
// For example, if Glasgow Airport's IATA code was in the CSV with the wrong city, it would be corrected here.
var CorrectCityMap = map[string]string{
	// Assuming "GLA" is the IATA code for Glasgow Airport, and the CSV inaccurately lists Paisley as its city.
	"GLA": "Glasgow",
	// Add other corrections here as needed.
}

func main() {
	// Open the database connection
	fmt.Println("Note: the default is to create a flights.db in this folder. If you are absolutely sure you want to update the data/flights.db then modify the main.go 'sql.open' lines and recompile")
	db, err := sql.Open("sqlite3", "./flights.db")
	//	db, err := sql.Open("sqlite3", "../../../../data/flights.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the airport_info table if it doesn't exist, updated to reflect new CSV structure
	createTableSQL := `CREATE TABLE IF NOT EXISTS airport_info (
		icao TEXT PRIMARY KEY,
		iata TEXT,
		name TEXT,
		city TEXT,
		subd TEXT,
		country TEXT,
		elevation INTEGER,
		lat REAL,
		lon REAL,
		tz TEXT,
		lid TEXT
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Open and read the CSV file
	csvFile, err := os.Open("airports.csv") // Update this path to your actual CSV file
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	reader.Comma = ',' // default, but being explicit
	reader.TrimLeadingSpace = true
	_, err = reader.Read() // Read and discard the header
	if err != nil {
		log.Fatal(err)
	}

	// Prepare the insert statement outside the loop for efficiency
	insertSQL := `INSERT OR REPLACE INTO airport_info (
		icao, iata, name, city, subd, country, elevation, lat, lon, tz, lid
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Read through the remaining records
	for {
		record, err := reader.Read()
		if err != nil {
			break // End of file or an error occurred
		}

		// Check if the city needs to be corrected based on the IATA code
		correctCity, exists := CorrectCityMap[record[1]]
		if exists {
			record[3] = correctCity // Update the city with the correct value
		}

		// Convert fields where necessary and insert
		_, err = stmt.Exec(
			record[0], record[1], record[2], record[3], record[4],
			record[5], record[6], record[7], record[8], record[9], record[10],
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Data successfully inserted into airport_info table.")
}
