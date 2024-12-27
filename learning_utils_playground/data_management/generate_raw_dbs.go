package data_management

// init-flights.go

import (
	"database/sql"
	"fmt"
	"log"

	db_manager "github.com/Tris20/FairFareFinder/learning_utils_playground/database_manager"
	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

// expected flights.db ???
func InitFlights(db *sql.DB) {
	rawDBFlight := db_manager.RawDBFlight{}
	_, err := db.Exec(rawDBFlight.CreateTableQuery())
	if err != nil {
		log.Fatal(err)
	}

	skyScannerPrice := db_manager.SkyScannerPrice{}
	_, err = db.Exec(skyScannerPrice.CreateTableQuery())
	if err != nil {
		log.Fatal(err)
	}
}

// init-locations-db.go
// expects locations.db
func InitLocationsDB(db *sql.DB) {
	// Create the 'city' table.
	locationsCity := db_manager.LocationsCity{}
	_, err := db.Exec(locationsCity.CreateTableQuery())
	if err != nil {
		log.Fatal(err)
	}

	// Create the 'airport' table.
	locationsAiport := db_manager.LocationsAirport{}
	_, err = db.Exec(locationsAiport.CreateTableQuery())
	if err != nil {
		log.Fatal(err)
	}

	// Create the 'marina' table.
	locationsMarina := db_manager.LocationsMarina{}
	_, err = db.Exec(locationsMarina.CreateTableQuery())
	if err != nil {
		log.Fatal(err)
	}

	// Create the 'beach' table.
	locationsBeach := db_manager.LocationsBeach{}
	_, err = db.Exec(locationsBeach.CreateTableQuery())
	if err != nil {
		log.Fatal(err)
	}

	// Create the 'ski_resort' table.
	locationsSkiResort := db_manager.LocationsSkiResort{}
	_, err = db.Exec(locationsSkiResort.CreateTableQuery())
	if err != nil {
		log.Fatal(err)
	}

	// Create the 'national_park' table.
	locationsNationalPark := db_manager.LocationsNationalPark{}
	_, err = db.Exec(locationsNationalPark.CreateTableQuery())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("All tables created successfully.")
}

// init-weather-db.go
// expects weather.db
func InitWeatherDb(db *sql.DB) {
	weatherRecord := db_manager.WeatherRecord{}
	_, err := db.Exec(weatherRecord.CreateTableQuery())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Weather table created successfully.")
}

// set-iata-cities-to-true.go
// expects locations.db
func SetIataCitiesToTrue(db *sql.DB) {
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
