
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
		INSERT INTO weather (city_name, country_code, date, weather_type, temperature, weather_icon_url, google_weather_link)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Create a new progress bar
	bar := progressbar.Default(int64(len(records)))

	for _, record := range records {
		_, err := stmt.Exec(record.CityName, record.CountryCode, record.Date, record.WeatherType, record.Temperature, record.WeatherIconURL, record.GoogleWeatherLink)
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

// InsertLocationData inserts unique location records into the location table in results.db
func InsertLocationData(destinationDBPath string, records []WeatherRecord) error {
	db, err := sql.Open("sqlite3", destinationDBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Extract unique locations
	uniqueLocations := getUniqueLocations(records)

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Prepare the insert statement
	stmt, err := tx.Prepare(`
		INSERT INTO location (city_name, country_code, iata, airbnb_url, booking_url, things_to_do, five_day_wpi)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Create a new progress bar
	bar := progressbar.Default(int64(len(uniqueLocations)))

	for _, loc := range uniqueLocations {
		// Generate URLs for flights and hotels
		skyScannerURL := GenerateSkyScannerURL(loc.IATA)
		airbnbURL := GenerateAirbnbURL(loc.CityName)
		bookingURL := GenerateBookingURL(loc.CityName)

		loc.SkyScannerURL = skyScannerURL
		loc.AirbnbURL = airbnbURL
		loc.BookingURL = bookingURL

		_, err := stmt.Exec(loc.CityName, loc.CountryCode, loc.IATA, loc.AirbnbURL, loc.BookingURL, loc.ThingsToDo, loc.FiveDayWPI)
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

// getUniqueLocations returns a list of unique locations from the given weather records, maintaining order
func getUniqueLocations(records []WeatherRecord) []Location {
	uniqueMap := make(map[string]struct{})
	var uniqueLocations []Location

	for _, record := range records {
		key := record.CityName + record.CountryCode
		if _, exists := uniqueMap[key]; !exists {
			uniqueMap[key] = struct{}{}
			uniqueLocations = append(uniqueLocations, Location{
				CityName:    record.CityName,
				CountryCode: record.CountryCode,
				IATA:        record.CityName, // Assuming IATA is same as city_name for simplicity
				AirbnbURL:   "placeholder_airbnb_url",
				BookingURL:  "placeholder_booking_url",
				ThingsToDo:  "placeholder_things_to_do",
				FiveDayWPI:  0.0,
			})
		}
	}

	return uniqueLocations
}

// Location struct to hold unique location data
type Location struct {
	CityName      string
	CountryCode   string
	IATA          string
	AirbnbURL     string
	BookingURL    string
	ThingsToDo    string
	FiveDayWPI    float64
	SkyScannerURL string // Added field for SkyScanner URL
}

