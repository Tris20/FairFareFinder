

package table_compilers

import (
	"database/sql"
  "fmt"
	"github.com/schollz/progressbar/v3"
	_ "github.com/mattn/go-sqlite3"

  "github.com/Tris20/FairFareFinder/src/backend" //types.go
)




func InsertLocationData(destinationDBPath string, uniqueLocations []Location) error {
	db, err := sql.Open("sqlite3", destinationDBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Prepare the insert statement
	stmt, err := tx.Prepare(`
		INSERT INTO location (city_name, country_code, iata, airbnb_url, booking_url, skyscanner_id, things_to_do, five_day_wpi)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Create a new progress bar
	bar := progressbar.Default(int64(len(uniqueLocations)))

	for _, loc := range uniqueLocations {
		_, err = stmt.Exec(loc.CityName, loc.CountryCode, loc.IATA, loc.AirbnbURL, loc.BookingURL, loc.SkyScannerID, loc.ThingsToDo, loc.FiveDayWPI)
		if err != nil {
			// Rollback the transaction in case of an error
			tx.Rollback()
			return err
		}
		// Increment the progress bar
		bar.Add(1)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}


