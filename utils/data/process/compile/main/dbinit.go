
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
		log.Fatalf("Failed to create Locations table: %v", err)
	}

	// Create Weather table
	createWeatherDetailedTable := `
	CREATE TABLE IF NOT EXISTS weather_detailed (
		weather_id INTEGER PRIMARY KEY AUTOINCREMENT,
		city_name VARCHAR(255),
		country_code CHAR(2),
		date DATE,
		weather_type VARCHAR(50),
		temperature DECIMAL(5,2),
		weather_icon_url VARCHAR(255),
		google_weather_link VARCHAR(255),
		wind_speed DECIMAL(5,2),
    wpi DECIMAL(5,2)
	);`
	_, err = db.Exec(createWeatherDetailedTable)
	if err != nil {
		log.Fatalf("Failed to create Weather table: %v", err)
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
}

