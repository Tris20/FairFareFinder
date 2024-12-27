package data_management

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

// CorrectCityMap provides a mapping for airports where the city information might need correction.
var CorrectCityMap = map[string]string{
	"GLA": "Glasgow", // Example correction
	// Add other corrections as needed.
}

func FetchLocationsAirports() {

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

// some break goes here

// updateSkyscannerID updates the Skyscanner ID for the given IATA code in the database
func updateSkyscannerID(db *sql.DB, skyscannerId, iata string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE airport SET skyscannerid = ? WHERE iata = ?", skyscannerId, iata)
	if err != nil {
		tx.Rollback() // Rollback in case of error
		return err
	}
	return tx.Commit() // Commit if all is good
}

// AddSkyScannerAirportIDs updates the Skyscanner IDs for airports in the database
func AddSkyScannerAirportIDs() {
	apiKey, err := readAPIKey("../../../../../ignore/secrets.yaml")
	if err != nil {
		log.Fatalf("Error reading API key: %v", err)
	}
	fmt.Println("Using API Key:", apiKey)

	// Connect to the SQLite database
	db, err := sql.Open("sqlite3", "./airports.db")
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Set journal_mode to WAL for better concurrency
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		fmt.Println("Failed to set PRAGMA journal_mode=WAL:", err)
		return
	}

	// Count the total number of records to update
	var totalCount int
	err = db.QueryRow("SELECT COUNT(*) FROM airport WHERE iata IS NOT NULL AND iata <> '' AND skyscannerid IS NULL").Scan(&totalCount)
	if err != nil {
		log.Fatalf("Error counting records: %v", err)
	}

	// Initialize the progress bar
	bar := progressbar.Default(int64(totalCount))

	// Query for IATA codes
	rows, err := db.Query("SELECT iata FROM airport WHERE iata IS NOT NULL AND iata <> '' AND skyscannerid IS NULL")
	if err != nil {
		log.Fatalf("Error querying database: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var iata string
		err = rows.Scan(&iata)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// API request to get Skyscanner ID
		url := fmt.Sprintf("https://skyscanner80.p.rapidapi.com/api/v1/flights/auto-complete?query=%s&market=US&locale=en-US", iata)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("X-RapidAPI-Key", apiKey)
		req.Header.Add("X-RapidAPI-Host", "skyscanner80.p.rapidapi.com")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Failed to make API request for IATA %s: %v", iata, err)
			continue
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Printf("Failed to read response body for IATA %s: %v", iata, err)
			continue
		}

		// Parse the JSON response
		var response ApiResponse_Airports
		json.Unmarshal(body, &response)

		// Update the database
		if len(response.Data) > 0 {
			skyscannerId := response.Data[0].Id
			err := updateSkyscannerID(db, skyscannerId, iata)
			if err != nil {
				log.Printf("Failed to update database for IATA %s: %v", iata, err)
				continue
			}
		} else {
			// Handle case where no Skyscanner ID is found
			err := updateSkyscannerID(db, "None", iata)
			if err != nil {
				log.Printf("Failed to update database for IATA %s with no Skyscanner ID: %v", iata, err)
				continue
			}
		}

		// Update progress bar
		bar.Add(1)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating through rows: %v", err)
	}
	fmt.Println("Successfully updated Skyscanner IDs for airports.")
}
