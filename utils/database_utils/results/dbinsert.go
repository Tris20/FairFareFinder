
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

	insertQuery := `
		INSERT INTO weather (city_name, country_code, date, weather_type, temperature, weather_icon_url, google_weather_link)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	// Create a new progress bar
	bar := progressbar.Default(int64(len(records)))

	for _, record := range records {
		_, err := db.Exec(insertQuery, record.CityName, record.CountryCode, record.Date, record.WeatherType, record.Temperature, record.WeatherIconURL, record.GoogleWeatherLink)
		if err != nil {
			return err
		}
		// Increment the progress bar
		bar.Add(1)
	}

	return nil
}

