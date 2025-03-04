package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

// Prediction represents one row from the prediction table.
type Prediction struct {
	OriginCity            string
	OriginCountry         string
	OriginIATA            string
	OriginPopulation      int
	DestinationCity       string
	DestinationCountry    string
	DestinationIATA       string
	DestinationPopulation int
	RouteFrequency        int
	RouteClassification   string
	MostCommonAirline     string
	MostCommonAircraft    string
	SeatingCapacity       int
	DurationHourDotMins   string
	PredictedPrice        float64
}

// SkyScannerPrice represents one row from the skyscannerprices table.
type SkyScannerPrice struct {
	OriginCity              string
	OriginCountry           string
	OriginIATA              string
	OriginSkyScannerID      string
	DestinationCity         string
	DestinationCountry      string
	DestinationIATA         string
	DestinationSkyScannerID string
	ThisWeekend             sql.NullFloat64
	NextWeekend             sql.NullFloat64
	SkyScannerURL           string
	SkyscannerDuration      sql.NullInt64
}

// buildSkyScannerURL builds a URL based on origin and destination IATA codes.
func buildSkyScannerURL(origin, dest string) string {
	return fmt.Sprintf("https://www.skyscanner.de/transport/fluge/%s/%s/?adults=1&adultsv2=1&cabinclass=economy&children=0&inboundaltsenabled=false&infants=0&outboundaltsenabled=false&preferdirects=true&ref=home&rtn=1", origin, dest)
}

func main() {
	fmt.Println("Starting to populate the main flights table...")

	// -----------------------------------
	// STEP 1: Read predictions from flight-prices.db (prediction table)
	// -----------------------------------
	predDB, err := sql.Open("sqlite3", "../../../../../../data/generated/flight-prices.db")
	if err != nil {
		log.Fatal("Error opening flight-prices.db: ", err)
	}
	defer predDB.Close()

	predRows, err := predDB.Query(`SELECT origin_city_name, origin_country, origin_iata,
		origin_population, destination_city_name, destination_country, destination_iata, destination_population,
		route_frequency, route_classification, most_common_airline, most_common_aircraft,
		most_common_aircraft_seating_capacity, duration_hour_dot_mins, predicted_price
		FROM prediction;`)
	if err != nil {
		log.Fatal("Error querying prediction table: ", err)
	}
	defer predRows.Close()

	var predictions []Prediction
	for predRows.Next() {
		var p Prediction
		err = predRows.Scan(&p.OriginCity, &p.OriginCountry, &p.OriginIATA, &p.OriginPopulation,
			&p.DestinationCity, &p.DestinationCountry, &p.DestinationIATA, &p.DestinationPopulation,
			&p.RouteFrequency, &p.RouteClassification, &p.MostCommonAirline, &p.MostCommonAircraft,
			&p.SeatingCapacity, &p.DurationHourDotMins, &p.PredictedPrice)
		if err != nil {
			log.Fatal("Error scanning prediction row: ", err)
		}
		predictions = append(predictions, p)
	}
	if err = predRows.Err(); err != nil {
		log.Fatal(err)
	}

	// -----------------------------------
	// STEP 2: Insert predictions into main flights table in new_main.db.
	// Also pull duration info from the flight-prices.db routes table.
	// -----------------------------------
	mainDB, err := sql.Open("sqlite3", "../../../../../../data/compiled/new_main.db")
	if err != nil {
		log.Fatal("Error opening new_main.db: ", err)
	}
	defer mainDB.Close()

	// Delete existing entries from the flight table.
	_, err = mainDB.Exec("DELETE FROM flight")
	if err != nil {
		log.Fatal("Failed to delete existing data: ", err)
	}
	fmt.Println("Existing data deleted from flight table.")

	// Create a progress bar for inserting predictions.
	bar := progressbar.Default(int64(len(predictions)), "Inserting predicted records")

	// The INSERT statement includes extra duration fields and skyscanner URLs.
	insertSQL := `INSERT INTO flight (
		origin_city_name, origin_country, origin_iata, origin_skyscanner_id,
		destination_city_name, destination_country, destination_iata, destination_skyscanner_id,
		price_this_week, skyscanner_url_this_week, price_next_week, skyscanner_url_next_week,
		duration_in_minutes, duration_in_hours, duration_in_hours_rounded, duration_hour_dot_mins
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	for _, p := range predictions {
		// For each prediction, look up the extra duration fields from the routes table in flight-prices.db.
		var durationMinutes int
		var durationHours float64
		var durationHoursRounded int
		var durationHdotMins sql.NullString
		err := predDB.QueryRow(`SELECT duration_in_minutes, duration_in_hours, duration_in_hours_rounded, duration_hour_dot_mins
			FROM routes
			WHERE origin_iata = ? AND destination_iata = ?`,
			p.OriginIATA, p.DestinationIATA).Scan(&durationMinutes, &durationHours, &durationHoursRounded, &durationHdotMins)
		if err != nil {
			// If not found, use defaults.
			durationMinutes = 0
			durationHours = 0
			durationHoursRounded = 0
			durationHdotMins.Valid = false
		}

		// Compute the default skyscanner URL using origin and destination IATA.
		url := buildSkyScannerURL(p.OriginIATA, p.DestinationIATA)

		// Insert predicted data along with the duration info and URL.
		_, err = mainDB.Exec(insertSQL,
			p.OriginCity, p.OriginCountry, p.OriginIATA, "", // origin_skyscanner_id empty for now
			p.DestinationCity, p.DestinationCountry, p.DestinationIATA, "", // destination_skyscanner_id empty for now
			p.PredictedPrice, url, // price_this_week, skyscanner_url_this_week
			p.PredictedPrice, url, // price_next_week, skyscanner_url_next_week
			durationMinutes, durationHours, durationHoursRounded,
			func() string {
				if durationHdotMins.Valid {
					return durationHdotMins.String
				}
				return ""
			}(),
		)
		if err != nil {
			log.Fatal("Error inserting prediction row: ", err)
		}
		bar.Add(1)
	}
	fmt.Println("\nPredicted records with duration info and skyscanner URL inserted into flight table of new_main.db.")

	// -----------------------------------
	// STEP 3: Overwrite predicted prices with skyscanner prices where available.
	// -----------------------------------
	// Open the raw flights database to read skyscannerprices.
	skyscannerDB, err := sql.Open("sqlite3", "../../../../../../data/raw/flights/flights.db")
	if err != nil {
		log.Fatal("Error opening raw flights database: ", err)
	}
	defer skyscannerDB.Close()

	skyscannerRows, err := skyscannerDB.Query(`SELECT origin_city, origin_country, origin_iata, origin_skyscanner_id,
		destination_city, destination_country, destination_iata, destination_skyscanner_id,
		this_weekend, next_weekend
		FROM skyscannerprices`)
	if err != nil {
		log.Fatal("Error querying skyscannerprices: ", err)
	}
	defer skyscannerRows.Close()

	var scannerPrices []SkyScannerPrice
	for skyscannerRows.Next() {
		var sp SkyScannerPrice
		err = skyscannerRows.Scan(&sp.OriginCity, &sp.OriginCountry, &sp.OriginIATA, &sp.OriginSkyScannerID,
			&sp.DestinationCity, &sp.DestinationCountry, &sp.DestinationIATA, &sp.DestinationSkyScannerID,
			&sp.ThisWeekend, &sp.NextWeekend)
		if err != nil {
			log.Fatal("Error scanning skyscannerprices row: ", err)
		}
		sp.OriginCountry = GetISOCode(sp.OriginCountry)
		sp.DestinationCountry = GetISOCode(sp.DestinationCountry)
		sp.SkyScannerURL = buildSkyScannerURL(sp.OriginIATA, sp.DestinationIATA)
		scannerPrices = append(scannerPrices, sp)
	}
	if err = skyscannerRows.Err(); err != nil {
		log.Fatal(err)
	}

	// For each skyscanner entry, update the corresponding row in flight table.
	// We match on origin_iata and destination_iata.
	for _, sp := range scannerPrices {
		var priceThisWeek, priceNextWeek float64
		if sp.ThisWeekend.Valid {
			priceThisWeek = sp.ThisWeekend.Float64
		}
		if sp.NextWeekend.Valid {
			priceNextWeek = sp.NextWeekend.Float64
		}
		_, err := mainDB.Exec(`UPDATE flight 
			SET origin_skyscanner_id = ?,
			    destination_skyscanner_id = ?,
			    price_this_week = ?,
			    skyscanner_url_this_week = ?,
			    price_next_week = ?,
			    skyscanner_url_next_week = ?
			WHERE origin_iata = ? AND destination_iata = ?`,
			sp.OriginSkyScannerID, sp.DestinationSkyScannerID,
			priceThisWeek, sp.SkyScannerURL,
			priceNextWeek, sp.SkyScannerURL,
			sp.OriginIATA, sp.DestinationIATA)
		if err != nil {
			log.Printf("Error updating flight for route %s -> %s: %v", sp.OriginIATA, sp.DestinationIATA, err)
		}
	}

	fmt.Println("Skyscanner prices have been used to update the flight table in new_main.db.")
}
