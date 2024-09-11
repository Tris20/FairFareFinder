package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Connect to SQLite database
	db, err := sql.Open("sqlite3", "../../../../../../data/compiled/new_main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// SQL statements to create tables
	createTables := []string{
		`CREATE TABLE IF NOT EXISTS flight (
			origin_iata CHAR(3) NOT NULL,
			origin_city VARCHAR(255) NOT NULL,
			origin_country CHAR(2) NOT NULL,
			destination_iata CHAR(3) NOT NULL,
			destination_city VARCHAR(255) NOT NULL,
			destination_country CHAR(2) NOT NULL,
			this_weekend DECIMAL(10, 2) NOT NULL,
			next_weekend DECIMAL(10, 2) NOT NULL,
			skyscanner_url VARCHAR(255),
			duration_mins INT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS weather (
			city VARCHAR(255) NOT NULL,
			country CHAR(2) NOT NULL,
			date DATE NOT NULL,
			avg_temp FLOAT(10,1),
			weather_icon VARCHAR(255),
			google_url VARCHAR(255),
			wpi FLOAT(10,1) 
		);`,
		`CREATE TABLE IF NOT EXISTS accommodation (
			city VARCHAR(255) NOT NULL,
			country CHAR(2) NOT NULL,
			airbnb_url VARCHAR(255),
			booking_url VARCHAR(255),
			avg_pppn DECIMAL(10, 2) NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS location (
			city VARCHAR(255) NOT NULL,
			country CHAR(2) NOT NULL,
			iata CHAR(3) NOT NULL,
			avg_wpi FLOAT(10,1)
		);`,
	}

	// Execute each CREATE TABLE statement
	for _, stmt := range createTables {
		_, err := db.Exec(stmt)
		if err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}
	}

	log.Println("Database initialized and tables created successfully.")
}

