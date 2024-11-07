package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
	"io"
	"log"
	"os"
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
			city VARCHAR(255) NOT NULL,
			country CHAR(2) NOT NULL,
			iata_1 CHAR(3) NOT NULL,
iata_2 CHAR(3) ,
iata_3 CHAR(3) ,
iata_4 CHAR(3) ,
iata_5 CHAR(3) ,
iata_6 CHAR(3) ,
iata_7 CHAR(3) ,
			avg_wpi FLOAT(10,1)
		);`
	_, err = db.Exec(createLocationsTable)
	if err != nil {
		log.Fatalf("Failed to create location table: %v", err)
	}

	// Create Weather table
	createWeatherDailyAverageTable := `
	CREATE TABLE IF NOT EXISTS weather (
			city VARCHAR(255) NOT NULL,
			country CHAR(2) NOT NULL,
			date DATE NOT NULL,
			avg_daytime_temp FLOAT(10,1),
			weather_icon VARCHAR(255),
			google_url VARCHAR(255),
			avg_daytime_wpi FLOAT(10,1) 	);`
	_, err = db.Exec(createWeatherDailyAverageTable)
	if err != nil {
		log.Fatalf("Failed to create Weather table: %v", err)
	}

	// Create flight_prices table
	createFlightPricesTable := `
CREATE TABLE IF NOT EXISTS "flight" (
	"id"	INTEGER,
	"origin_city_name"	TEXT,
	"origin_country"	TEXT,
	"origin_iata"	TEXT,
	"origin_skyscanner_id"	TEXT,
	"destination_city_name"	TEXT,
	"destination_country"	TEXT,
	"destination_iata"	TEXT,
	"destination_skyscanner_id"	TEXT,
	"price_this_week"	DECIMAL,
	"skyscanner_url_this_week"	VARCHAR(255),
	"price_next_week"	DECIMAL,
	"skyscanner_url_next_week"	VARCHAR(255),
	"duration_in_minutes"	DECIMAL,
	PRIMARY KEY("id" AUTOINCREMENT)
);
`
	_, err = db.Exec(createFlightPricesTable)
	if err != nil {
		log.Fatalf("Failed to create flight_prices table: %v", err)
	}

	log.Println("Database and tables created successfully.")

	// Create five_nights_and_flights table
	createFiveNightsAndFlightsTable := `
CREATE TABLE IF NOT EXISTS five_nights_and_flights (
    origin_city TEXT,
    origin_country TEXT,
    destination_city TEXT,
    destination_country TEXT,
    price_fnaf REAL
);`
	_, err = db.Exec(createFiveNightsAndFlightsTable)
	if err != nil {
		log.Fatalf("Failed to create five_nights_and_flights table: %v", err)
	}

	// Create accommodation table
	createAccommodationTable := `
CREATE TABLE IF NOT EXISTS accommodation (
    city TEXT NOT NULL,
    country TEXT NOT NULL,
    booking_url TEXT,
    booking_pppn REAL NOT NULL
);`
	_, err = db.Exec(createAccommodationTable)
	if err != nil {
		log.Fatalf("Failed to create accommodation table: %v", err)
	}

}

// Helper function to delete new_main.db if it exists
func deleteNewMainDB(dbPath string) error {
	// Check if new_main.db exists
	if _, err := os.Stat(dbPath); err == nil {
		// Delete new_main.db
		err := os.Remove(dbPath)
		if err != nil {
			return fmt.Errorf("failed to delete new_main.db: %v", err)
		}
		fmt.Println("Deleted existing new_main.db")
	} else if !os.IsNotExist(err) {
		// Some other error occurred, but not a "file doesn't exist" error
		return fmt.Errorf("failed to check new_main.db: %v", err)
	}

	return nil
}

// Helper function to copy main.db to new_main.db
func copyMainDB(srcPath, destPath string) error {
	// Open source main.db
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open main.db: %v", err)
	}
	defer srcFile.Close()

	// Create destination new_main.db
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create new_main.db: %v", err)
	}
	defer destFile.Close()

	// Copy the content from main.db to new_main.db
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy main.db to new_main.db: %v", err)
	}

	fmt.Println("Successfully copied main.db to new_main.db")
	return nil
}
