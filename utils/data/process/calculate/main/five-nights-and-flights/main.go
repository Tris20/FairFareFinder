
package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

func main() {
	// Step 1: Open the database
	db, err := sql.Open("sqlite3", "../../../../../../data/compiled/new_main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Step 2: Check if the "five_nights_and_flights" table exists, and create it if not
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS five_nights_and_flights (
		origin_city TEXT,
		origin_country TEXT,
		destination_city TEXT,
		destination_country TEXT,
		price_fnaf REAL
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Table 'five_nights_and_flights' ensured to exist or created successfully.")

	// Step 3: Count the total rows to know the progress length
	var rowCount int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM flight
	`).Scan(&rowCount)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total number of rows to process: %d\n", rowCount)

	// Step 4: Query data from the "flight" table
	rows, err := db.Query(`
		SELECT origin_city_name, origin_country, destination_city_name, destination_country, price_this_week
		FROM flight
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Step 5: Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Step 6: Prepare the progress bar
	bar := progressbar.NewOptions(rowCount,
		progressbar.OptionSetDescription("Processing rows..."),
		progressbar.OptionShowCount(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowIts(),
		progressbar.OptionClearOnFinish(),
	)

	// Step 7: Prepare the insert statement for five_nights_and_flights table
	insertQuery := `
	INSERT INTO five_nights_and_flights (origin_city, origin_country, destination_city, destination_country, price_fnaf)
	VALUES (?, ?, ?, ?, ?)
	`
	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Step 8: Iterate over the rows from the "flight" table
	for rows.Next() {
		var originCity, originCountry, destCity, destCountry string
		var flightPrice float64

		// Scan the flight data
		err := rows.Scan(&originCity, &originCountry, &destCity, &destCountry, &flightPrice)
		if err != nil {
			log.Fatal(err)
		}

		// Step 9: Get booking_pppn from the "accommodation" table for the same destination
		var bookingPPPN float64
		err = db.QueryRow(`
			SELECT booking_pppn
			FROM accommodation
			WHERE city = ? AND country = ?
		`, destCity, destCountry).Scan(&bookingPPPN)

		if err != nil {
			// If there's no matching entry in the accommodation table, handle it gracefully
			if err == sql.ErrNoRows {
				bookingPPPN = 0 // No accommodation price available, set to 0
			} else {
				log.Fatal(err)
			}
		}

		// Step 10: Calculate the final price (flight + 5 nights of accommodation)
		totalPriceFNAF := flightPrice + (bookingPPPN * 5)

		// Step 11: Insert the result into "five_nights_and_flights" table
		_, err = stmt.Exec(originCity, originCountry, destCity, destCountry, totalPriceFNAF)
		if err != nil {
			log.Fatal(err)
		}

		// Update the progress bar for each row processed
		bar.Add(1)
	}

	// Step 12: Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nData inserted into 'five_nights_and_flights' table successfully.")
}

