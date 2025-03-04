package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func CreateResultsDB() {
	// Ensure the "generated" directory exists.
	dir := "../../../../../data/generated"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	// Define the database file path.
	dbPath := filepath.Join(dir, "flight-prices.db")

	// Open the SQLite database. It will be created if it doesn't exist.
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the "routes" table with the specified columns.
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS routes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		origin_city_name TEXT,
		origin_country TEXT,
		origin_iata TEXT,
		origin_population INTEGER,
		destination_city_name TEXT,
		destination_country TEXT,
		destination_iata TEXT,
		destination_population INTEGER,
		route_frequency INTEGER,
		route_classification TEXT,
		most_common_airline TEXT,
		most_common_aircraft TEXT,
		most_common_aircraft_seating_capacity INTEGER,
		duration_in_minutes INTEGER,
		duration_in_hours REAL,
		duration_in_hours_rounded INTEGER,
		duration_hour_dot_mins TEXT,
		calculated_price REAL
	);
	`

	// Execute the table creation query.
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	fmt.Println("Database and table 'routes' created successfully in", dbPath)
}
