package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// backupDatabase checks if the database file exists and, if so, renames it with a timestamp.
func backupDatabase(dbFilePath string) error {
	// Check if the file exists.
	if _, err := os.Stat(dbFilePath); err == nil {
		// File exists; create a backup file name using the current date.
		timestamp := time.Now().Format("2006_01_02")
		backupFile := fmt.Sprintf("flight_price_modifiers_backup_%s.db", timestamp)

		// Ensure backup is in the same directory as the original file.
		dir := filepath.Dir(dbFilePath)
		backupPath := filepath.Join(dir, backupFile)

		err = os.Rename(dbFilePath, backupPath)
		if err != nil {
			return fmt.Errorf("failed to backup database: %w", err)
		}
		log.Printf("Database backed up to %s", backupPath)
	} else if os.IsNotExist(err) {
		// File does not exist; nothing to do.
		log.Println("No existing database found; no backup needed.")
	} else {
		// Some other error occurred while checking the file.
		return fmt.Errorf("error checking for database file: %w", err)
	}

	return nil
}

func main() {
	// Path to your database file.
	dbFilePath := "../../../data/generated/flight_price_modifiers.db"

	// Backup the existing database (if it exists) using the current date.
	if err := backupDatabase(dbFilePath); err != nil {
		log.Fatalf("Error backing up database: %v", err)
	}

	// Now open (or create) the new SQLite database.
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// Create the airline multipliers table
	createAirlineTableSQL := `
	CREATE TABLE IF NOT EXISTS airline_multipliers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		airline TEXT NOT NULL UNIQUE,
		multiplier REAL NOT NULL
	);`
	_, err = db.Exec(createAirlineTableSQL)
	if err != nil {
		log.Fatalf("Error creating airline_multipliers table: %v", err)
	}

	// Create the date (season/holiday) modifiers table
	createDateModifiersSQL := `
	CREATE TABLE IF NOT EXISTS date_modifiers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		start_date TEXT NOT NULL, 
		end_date TEXT NOT NULL,  
		multiplier REAL NOT NULL,
		reason TEXT,
		countries TEXT  
	);`
	_, err = db.Exec(createDateModifiersSQL)
	if err != nil {
		log.Fatalf("Error creating date_modifiers table: %v", err)
	}

	// Create the population modifiers table
	createPopulationModifiersSQL := `
	CREATE TABLE IF NOT EXISTS population_modifiers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		min_population INTEGER NOT NULL,
		max_population INTEGER NOT NULL,
		multiplier REAL NOT NULL,
		description TEXT
	);`
	_, err = db.Exec(createPopulationModifiersSQL)
	if err != nil {
		log.Fatalf("Error creating population_modifiers table: %v", err)
	}

	// Create the flight frequency modifiers table
	createFlightFrequencyModifiersSQL := `
	CREATE TABLE IF NOT EXISTS flight_frequency_modifiers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		min_flights INTEGER NOT NULL,
		max_flights INTEGER, 
		multiplier REAL NOT NULL,
		notes TEXT
	);`
	_, err = db.Exec(createFlightFrequencyModifiersSQL)
	if err != nil {
		log.Fatalf("Error creating flight_frequency_modifiers table: %v", err)
	}

	// Create the short-notice modifiers table
	createShortNoticeModifiersSQL := `
	CREATE TABLE IF NOT EXISTS short_notice_modifiers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		min_days INTEGER NOT NULL,
		max_days INTEGER, 
		multiplier REAL NOT NULL,
		explanation TEXT
	);`
	_, err = db.Exec(createShortNoticeModifiersSQL)
	if err != nil {
		log.Fatalf("Error creating short_notice_modifiers table: %v", err)
	}

	// Create the aircraft capacity modifiers table
	createAircraftCapacityModifiersSQL := `
	CREATE TABLE IF NOT EXISTS aircraft_capacity_modifiers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		min_capacity INTEGER NOT NULL,
		max_capacity INTEGER, 
		multiplier REAL NOT NULL,
		description TEXT
	);`
	_, err = db.Exec(createAircraftCapacityModifiersSQL)
	if err != nil {
		log.Fatalf("Error creating aircraft_capacity_modifiers table: %v", err)
	}

	// Create the route classification modifiers table
	createRouteClassificationModifiersSQL := `
CREATE TABLE IF NOT EXISTS route_classification_modifiers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    classification TEXT NOT NULL, 
    multiplier REAL NOT NULL,       
    description TEXT
);`
	_, err = db.Exec(createRouteClassificationModifiersSQL)
	if err != nil {
		log.Fatalf("Error creating route_classification_modifiers table: %v", err)
	}

	log.Println("Database and all tables created successfully!")

	// Populate the tables with data.
	if err := populateTables(db); err != nil {
		log.Fatalf("Error populating tables: %v", err)
	}

	log.Println("Database populated with multiplier data successfully!")
}
