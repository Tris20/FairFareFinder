package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	//	"time"
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

func updateSkyscannerID(db *sql.DB, skyscannerId, iata string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE airport_info SET skyscannerid = ? WHERE iata = ?", skyscannerId, iata)
	if err != nil {
		tx.Rollback() // Rollback in case of error
		return err
	}
	err = tx.Commit() // Commit if all is good
	return err
}
func main() {

	apiKey, err := readAPIKey("../../../../ignore/secrets.yaml")
	if err != nil {
		log.Fatalf("Error reading API key: %v", err)
	}
	fmt.Println(apiKey)
	// Connect to the SQLite database
	db, err := sql.Open("sqlite3", "./flights.db")
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}
	defer db.Close()

	// Set journal_mode to WAL for better concurrency
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		fmt.Println("Failed to set PRAGMA journal_mode=WAL:", err)
		return
	}

	rows, err := db.Query("SELECT iata FROM airport_info WHERE iata IS NOT NULL AND iata <> '' AND skyscannerid IS NULL")
	if err != nil {
		fmt.Println("Error querying database:", err)
		return
	}
	defer rows.Close()

	var count int
	// Iterate through the result set
	for rows.Next() {
		if count < 3999 {
			count++
			var iata string
			err = rows.Scan(&iata)
			if err != nil {
				fmt.Println("Error scanning row:", err)
				continue
			}

			// Perform the API request with the current IATA code
			url := fmt.Sprintf("https://skyscanner80.p.rapidapi.com/api/v1/flights/auto-complete?query=%s&market=US&locale=en-US", iata)
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Add("X-RapidAPI-Key", apiKey)
			req.Header.Add("X-RapidAPI-Host", "skyscanner80.p.rapidapi.com")

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println("Failed to make API request:", err)
				continue
			}
			body, err := io.ReadAll(res.Body)
			res.Body.Close() // Close immediately after reading
			if err != nil {
				fmt.Println("Failed to read response body:", err)
				continue
			} // Print the API response
			fmt.Println("API Response for IATA:", iata)
			fmt.Println(string(body))

			// Parse the JSON response
			var response ApiResponse
			json.Unmarshal(body, &response)

			// Update the database with the Skyscanner ID
			if len(response.Data) > 0 {
				skyscannerId := fmt.Sprintf(response.Data[0].Id)
				fmt.Println(skyscannerId)

				if err := updateSkyscannerID(db, skyscannerId, iata); err != nil {
					fmt.Println("Failed to update database after retries:", err)
					continue
				}
			} else {

				skyscannerId := "None"

				if err := updateSkyscannerID(db, skyscannerId, iata); err != nil {
					fmt.Println("Failed to update database after retries:", err)
					continue
				}
			}
		}
	}
	if err := rows.Err(); err != nil {
		fmt.Println("Error iterating through rows:", err)
	}
}
