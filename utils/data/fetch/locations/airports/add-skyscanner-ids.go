
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

// ApiResponse defines the structure to parse the JSON response
type ApiResponse struct {
	Data []struct {
		Id string `json:"id"`
	} `json:"data"`
}

// Secrets struct to match the YAML structure
type Secrets struct {
	APIKeys struct {
		SkyScanner string `yaml:"skyscanner"`
	} `yaml:"api_keys"`
}

// readAPIKey reads the API key from the provided YAML file path
func readAPIKey(filepath string) (string, error) {
	var secrets Secrets
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	err = yaml.Unmarshal(file, &secrets)
	if err != nil {
		return "", err
	}
	return secrets.APIKeys.SkyScanner, nil
}

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
		var response ApiResponse
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


