
package main

import (
	"database/sql"

	"github.com/schollz/progressbar/v3"
	_ "github.com/mattn/go-sqlite3"
)

// InsertWeatherData inserts weather records into the weather table in results.db
func InsertWeatherData(destinationDBPath string, records []WeatherRecord) error {
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
		INSERT INTO weather (city_name, country_code, date, weather_type, temperature, weather_icon_url, google_weather_link, wind_speed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Create a new progress bar
	bar := progressbar.Default(int64(len(records)))

	for _, record := range records {
		_, err := stmt.Exec(record.CityName, record.CountryCode, record.Date, record.WeatherType, record.Temperature, record.WeatherIconURL, record.GoogleWeatherLink, record.WindSpeed)
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

