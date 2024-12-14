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
		`CREATE TABLE IF NOT EXISTS "flight" (
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
  "duration_in_hours"	DECIMAL,
  "duration_hour_dot_mins" REAL,
	PRIMARY KEY("id" AUTOINCREMENT)
	);`,
		`CREATE TABLE IF NOT EXISTS weather (
			city VARCHAR(255) NOT NULL,
			country CHAR(2) NOT NULL,
			date DATE NOT NULL,
			avg_daytime_temp FLOAT(10,1),
			weather_icon VARCHAR(255),
			google_url VARCHAR(255),
			avg_daytime_wpi FLOAT(10,1) 
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
			iata_1 CHAR(3) NOT NULL,
iata_2 CHAR(3) ,
iata_3 CHAR(3) ,
iata_4 CHAR(3) ,
iata_5 CHAR(3) ,
iata_6 CHAR(3) ,
iata_7 CHAR(3) ,
			avg_wpi FLOAT(10,1),
      image_1 TEXT,
      image_2 TEXT,
      image_3 TEXT,
      image_4 TEXT,
      image_5 TEXT,
      image_6 TEXT,
      image_7 TEXT,
      image_8 TEXT,
      image_9 TEXT,
      UNIQUE(city, country)
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
