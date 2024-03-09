package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// AirportInfo represents the structure of our JSON data
type AirportInfo struct {
	Continent    string  `json:"continent"`
	Coordinates  string  `json:"coordinates"`
	ElevationFt  string  `json:"elevation_ft"`
	GpsCode      *string `json:"gps_code"`
	IataCode     *string `json:"iata_code"`
	Ident        string  `json:"ident"`
	IsoCountry   string  `json:"iso_country"`
	IsoRegion    string  `json:"iso_region"`
	LocalCode    *string `json:"local_code"`
	Municipality *string `json:"municipality"`
	Name         string  `json:"name"`
	Type         string  `json:"type"`
}

func main() {
	// Open the database connection
	db, err := sql.Open("sqlite3", "../../../data/flights.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the airport_info table
	createTableSQL := `CREATE TABLE IF NOT EXISTS airport_info (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		continent TEXT,
		coordinates TEXT,
		elevation_ft TEXT,
		gps_code TEXT,
		iata_code TEXT,
		ident TEXT NOT NULL,
		iso_country TEXT,
		iso_region TEXT,
		local_code TEXT,
		municipality TEXT,
		name TEXT,
		type TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Open and read the JSON file
	jsonFile, err := os.ReadFile("detailed-airport-info.json") // Replace with your JSON file's path
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal the JSON data into a slice of AirportInfo structs
	var airports []AirportInfo
	err = json.Unmarshal(jsonFile, &airports)
	if err != nil {
		log.Fatal(err)
	}

	// Insert the data into the airport_info table
	for _, airport := range airports {
		insertSQL := `INSERT INTO airport_info (
			continent, coordinates, elevation_ft, gps_code, iata_code,
			ident, iso_country, iso_region, local_code, municipality,
			name, type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		_, err = db.Exec(insertSQL, airport.Continent, airport.Coordinates, airport.ElevationFt, airport.GpsCode, airport.IataCode,
			airport.Ident, airport.IsoCountry, airport.IsoRegion, airport.LocalCode, airport.Municipality,
			airport.Name, airport.Type)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Data successfully inserted into airport_info table.")
}

