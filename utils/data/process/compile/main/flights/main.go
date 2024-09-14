
package main

import (
	"database/sql"
	"fmt"
	"log"
	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

type SkyScannerPrice struct {
	OriginCity            string
	OriginCountry         string
	OriginIATA            string
	OriginSkyScannerID    string
	DestinationCity       string
	DestinationCountry    string
	DestinationIATA       string
	DestinationSkyScannerID string
	ThisWeekend           sql.NullFloat64
	NextWeekend           sql.NullFloat64
	SkyScannerURL         string
}



func main() {
	db, err := sql.Open("sqlite3", "../../../../../../data/raw/flights/flights.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM skyscannerprices")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var prices []SkyScannerPrice

	for rows.Next() {
		var sp SkyScannerPrice
		err = rows.Scan(&sp.OriginCity, &sp.OriginCountry, &sp.OriginIATA, &sp.OriginSkyScannerID, &sp.DestinationCity, &sp.DestinationCountry, &sp.DestinationIATA, &sp.DestinationSkyScannerID, &sp.ThisWeekend, &sp.NextWeekend)
		if err != nil {
			log.Fatal(err)
		}
		sp.OriginCountry = GetISOCode(sp.OriginCountry)
		sp.SkyScannerURL = fmt.Sprintf("https://www.skyscanner.de/transport/fluge/%s/%s/?adults=1&adultsv2=1&cabinclass=economy&children=0&inboundaltsenabled=false&infants=0&outboundaltsenabled=false&preferdirects=true&ref=home&rtn=1", sp.OriginIATA, sp.DestinationIATA)
		prices = append(prices, sp)
	}
	if rows.Err() != nil {
		log.Fatal(rows.Err())
	}

	// Open a new connection to the `new_main.db` database
	newDB, err := sql.Open("sqlite3", "../../../../../../data/compiled/new_main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer newDB.Close()

	// Delete all existing entries from the flight table
	_, err = newDB.Exec("DELETE FROM flight")
	if err != nil {
		log.Fatal("Failed to delete existing data: ", err)
	}
	fmt.Println("Existing data deleted.")

	// Create a new progress bar
	bar := progressbar.Default(int64(len(prices)), "Inserting records")

	// Insert data into the new table
	for _, price := range prices {
		_, err := newDB.Exec("INSERT INTO flight (origin_city_name, origin_country, origin_iata, origin_skyscanner_id, destination_city_name, destination_country, destination_iata, destination_skyscanner_id, price_this_week, skyscanner_url_this_week, price_next_week, skyscanner_url_next_week) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			price.OriginCity, price.OriginCountry, price.OriginIATA, price.OriginSkyScannerID, price.DestinationCity, price.DestinationCountry, price.DestinationIATA, price.DestinationSkyScannerID, price.ThisWeekend.Float64, price.SkyScannerURL, price.NextWeekend.Float64, price.SkyScannerURL)
		if err != nil {
			log.Fatal(err)
		}
		bar.Add(1) // Update the progress bar
	}
	fmt.Println("\nData inserted successfully into new_main.db")
}

