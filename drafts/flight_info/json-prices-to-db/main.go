package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Structure to hold price information
type PriceInfo struct {
	Price float64 `json:"Price"`
}

func main() {
	// Load JSON data from file
	jsonData, err := ioutil.ReadFile("prices.json")
	if err != nil {
		log.Fatalf("failed to read prices.json: %v", err)
	}

	// Parse JSON data
	var prices map[string]PriceInfo
	err = json.Unmarshal(jsonData, &prices)
	if err != nil {
		log.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", "../../../data/flights.db")
	if err != nil {
		log.Fatalf("failed to open flights.db: %v", err)
	}
	defer db.Close()

	// Create table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS skyscannerprices (
		origin TEXT,
		destination TEXT,
		this_weekend REAL,
		next_weekend REAL
	)`)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	// Insert data into table
	for key, value := range prices {
		parts := strings.SplitN(key, "_", 2)
		if len(parts) < 2 {
			continue // Skip invalid entries
		}
		origin := parts[0]
		destination := parts[1]

		_, err := db.Exec("INSERT INTO skyscannerprices (origin, destination, this_weekend, next_weekend) VALUES (?, ?, ?, ?)",
			origin, destination, value.Price, nil)
		if err != nil {
			log.Printf("failed to insert data: %v", err)
			continue
		}
	}

	fmt.Println("Data inserted successfully.")
}
