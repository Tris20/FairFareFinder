
package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

func initializeDatabase(dbPath string) {
	// Open the database file
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create Locations table
	createLocationsTable := `
	CREATE TABLE IF NOT EXISTS location (
		location_id INTEGER PRIMARY KEY AUTOINCREMENT,
    city_name VARCHAR(255),
		airbnb_url VARCHAR(255),
		booking_url VARCHAR(255),
		things_to_do TEXT
	);`
	_, err = db.Exec(createLocationsTable)
	if err != nil {
		log.Fatalf("Failed to create Locations table: %v", err)
	}

	// Create Weather table
	createWeatherTable := `
	CREATE TABLE IF NOT EXISTS weather (
		weather_id INTEGER PRIMARY KEY AUTOINCREMENT,
		city_name VARCHAR(255),
		country_code CHAR(2),
		date DATE,
		weather_type VARCHAR(50),
		temperature DECIMAL(5,2),
		wpi DECIMAL(2),
		weather_icon_url VARCHAR(255),
		google_weather_link VARCHAR(255)
	);`
	_, err = db.Exec(createWeatherTable)
	if err != nil {
		log.Fatalf("Failed to create Weather table: %v", err)
	}

	// Create flight_prices table
	createFlightPricesTable := `
	CREATE TABLE IF NOT EXISTS flight_price (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		city_name TEXT,
		destination TEXT,
		price_this_week DECIMAL,
		price_next_week DECIMAL
	);`
	_, err = db.Exec(createFlightPricesTable)
	if err != nil {
		log.Fatalf("Failed to create flight_prices table: %v", err)
	}

	log.Println("Database and tables created successfully.")
}

