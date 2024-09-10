
package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open the SQLite database.
	db, err := sql.Open("sqlite3", "../../../../../../data/raw/flights/flights.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE "schedule" (
	"id"	INTEGER,
	"flightNumber"	TEXT NOT NULL,
	"departureAirport"	TEXT,
	"arrivalAirport"	TEXT,
	"departureTime"	TEXT,
	"arrivalTime"	TEXT,
	"direction"	TEXT NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT)
);`)

if err != nil {
		log.Fatal(err)
	}
log.Println("schedule table created successfully.")


	_, err = db.Exec(`CREATE TABLE "skyscannerprices" (
	"origin_iata"	TEXT,
  "origin"	TEXT,
	"destination_iata"	TEXT,
  "destination"	TEXT,
	"this_weekend"	REAL,
	"next_weekend"	REAL
);`)
	if err != nil {
		log.Fatal(err)
	}


	log.Println("skyscannerprices table created successfully.")
}
