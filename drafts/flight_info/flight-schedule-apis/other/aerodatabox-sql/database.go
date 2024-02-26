
package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// InitializeDB sets up the SQLite database and the flights table
func InitializeDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS flights (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"flightType" TEXT,
		"departureAirport" TEXT,
		"departureTime" TEXT,
		"arrivalAirport" TEXT,
		"arrivalTime" TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

// InsertFlightData adds a new flight record to the database
func InsertFlightData(db *sql.DB, flightType, departureAirport, departureTime, arrivalAirport, arrivalTime string) {
    insertSQL := `INSERT INTO flights (flightType, departureAirport, departureTime, arrivalAirport, arrivalTime) VALUES (?, ?, ?, ?, ?)`	
    stmt, err := db.Prepare(insertSQL)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(flightType, departureAirport, departureTime, arrivalAirport, arrivalTime)
	if err != nil {
		log.Fatal(err)
	}
}
