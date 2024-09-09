
package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open the SQLite database.
	db, err := sql.Open("sqlite3", "../../../../../../data/raw/locations/locations.db")
	if err != nil {
		log.Fatal(err)
	}
  defer db.Close()

	// Create the 'city' table.
	_, err = db.Exec(`CREATE TABLE city (
		id INTEGER PRIMARY KEY,
		include_tf BOOLEAN,
		city TEXT,
		countrycode TEXT,
		population BIGINT,
		elevation REAL,
		lat REAL,
		long REAL
	);`)
	if err != nil {
		log.Fatal(err)
	}

	// Create the 'airport' table.
	_, err = db.Exec(`CREATE TABLE "airport" (
		"icao" TEXT,
		"iata" TEXT,
		"name" TEXT,
		"city" TEXT,
		"subd" BLOB,
		"country" TEXT,
		"elevation" INTEGER,
		"lat" REAL,
		"lon" REAL,
		"tz" TEXT,
		"lid" TEXT,
		"skyscannerid" TEXT,
		PRIMARY KEY("icao")
	);`)
	if err != nil {
		log.Fatal(err)
	}

	// Create the 'marina' table.
	_, err = db.Exec(`CREATE TABLE marina (
		id INTEGER PRIMARY KEY,
		name TEXT,
		location TEXT,
		capacity INTEGER,
		facilities TEXT,
		lat REAL,
		lon REAL
	);`)
	if err != nil {
		log.Fatal(err)
	}

	// Create the 'beach' table.
	_, err = db.Exec(`CREATE TABLE beach (
		id INTEGER PRIMARY KEY,
		name TEXT,
		location TEXT,
		accessibility TEXT,
		facilities TEXT,
		water_quality TEXT,
		lat REAL,
		lon REAL
	);`)
	if err != nil {
		log.Fatal(err)
	}

	// Create the 'ski_resort' table.
	_, err = db.Exec(`CREATE TABLE ski_resort (
		id INTEGER PRIMARY KEY,
		name TEXT,
		location TEXT,
		num_trails INTEGER,
		difficulty TEXT,
		lift_count INTEGER,
		lat REAL,
		lon REAL
	);`)
	if err != nil {
		log.Fatal(err)
	}

	// Create the 'national_park' table.
	_, err = db.Exec(`CREATE TABLE national_park (
		id INTEGER PRIMARY KEY,
		name TEXT,
		location TEXT,
		area_sq_km REAL,
		visitors_per_year INTEGER,
		established_year INTEGER,
		lat REAL,
		lon REAL
	);`)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("All tables created successfully.")
}
