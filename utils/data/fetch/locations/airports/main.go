
package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"github.com/schollz/progressbar/v3"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

// CorrectCityMap provides a mapping for airports where the city information might need correction.
var CorrectCityMap = map[string]string{
	"GLA": "Glasgow", // Example correction
	// Add other corrections as needed.
}

func main() {
  
	fmt.Println("Note: the default is to create a airports.db in this folder. If you are absolutely sure you want to update the data/locations.db then modify the main.go 'sql.open' lines and recompile")
	db, err := sql.Open("sqlite3", "./airports.db")
//	db, err := sql.Open("sqlite3", "../../../../../data/raw/locations/locations.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTableSQL := `CREATE TABLE IF NOT EXISTS airport (
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
		lid TEXT,
    skyscannerid TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}



	csvFile, err := os.Open("airports.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	// Count total records for progress bar, then reset file pointer
	totalRecords, err := countRecords(csvFile)
	if err != nil {
		log.Fatal("Failed to count records:", err)
	}
	_, err = csvFile.Seek(0, 0) // Reset to beginning of file
	if err != nil {
		log.Fatal(err)
	}

	reader := csv.NewReader(csvFile)
	reader.Comma = ',' // Default, being explicit
	reader.TrimLeadingSpace = true
	_, err = reader.Read() // Read and discard the header
	if err != nil {
		log.Fatal(err)
	}

	bar := progressbar.Default(int64(totalRecords - 1)) // -1 to exclude header

	insertSQL := `INSERT OR REPLACE INTO airport_info (
		icao, iata, name, city, subd, country, elevation, lat, lon, tz, lid
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for {
		record, err := reader.Read()
		if err != nil {
			break // End of file or an error occurred
		}

		// Correct city if necessary
		if correctCity, exists := CorrectCityMap[record[1]]; exists {
			record[3] = correctCity
		}

		// Insert the record
		_, err = stmt.Exec(record[0], record[1], record[2], record[3], record[4], record[5], record[6], record[7], record[8], record[9], record[10])
		if err != nil {
			log.Fatal(err)
		}

		bar.Add(1)
	}

	fmt.Println("Data successfully inserted into airport_info table.")
  
  fmt.Println("\nAdding AddSkyScannerAirportIDs")
  AddSkyScannerAirportIDs()
}

// countRecords returns the total number of records in the CSV file, including the header.
func countRecords(file *os.File) (int, error) {
	// Reset to beginning of file to ensure accurate count
	_, err := file.Seek(0, 0)
	if err != nil {
		return 0, err
	}

	reader := csv.NewReader(file)
	count := 0
	for {
		_, err := reader.Read()
		if err != nil {
			break
		}
		count++
	}

	// No need to reset file pointer here, it will be reset in the main function
	return count, nil
}

