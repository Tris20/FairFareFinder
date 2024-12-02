package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

func main() {
	// Open the SQLite database.
	db, err := sql.Open("sqlite3", "../../../../../../../data/raw/locations/locations.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Begin a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Query to count the total rows first for the progress bar
	var totalRows int
	err = tx.QueryRow("SELECT COUNT(*) FROM airport WHERE iata IS NOT NULL").Scan(&totalRows)
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	// Create a new progress bar
	bar := progressbar.NewOptions(totalRows,
		progressbar.OptionSetDescription("Updating cities..."),
		progressbar.OptionSetRenderBlankState(true),
	)

	// Step 1: Get all rows from the "airport" table where the iata column is not null, converting to lowercase
	rows, err := tx.Query("SELECT LOWER(city), LOWER(country) FROM airport WHERE iata IS NOT NULL")
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}
	defer rows.Close()

	// Step 2: In the "city" table, set the "include_tf" row to 1 where conditions match, converting to lowercase
	stmt, err := tx.Prepare("UPDATE city SET include_tf = 1 WHERE LOWER(city_ascii) = ? AND LOWER(iso2) = ?")
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}
	defer stmt.Close()

	// Using sql.NullString to handle NULL values and updating progress bar
	var city, country sql.NullString
	for rows.Next() {
		err := rows.Scan(&city, &country)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
		if city.Valid && country.Valid { // Check if both city and country are not NULL
			_, err = stmt.Exec(city.String, country.String)
			if err != nil {
				tx.Rollback()
				log.Fatal(err)
			}
		}
		bar.Add(1) // Update the progress bar
	}
	if err = rows.Err(); err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("The 'include_tf' flags were updated successfully.")
}
