
package main

import (
	"database/sql"
	"time"
"fmt"
	_ "github.com/mattn/go-sqlite3"
)

// FetchWeatherData retrieves weather data for today and the next 5 days from weather.db
func FetchWeatherData(sourceDBPath string) ([]WeatherRecord, error) {
	db, err := sql.Open("sqlite3", sourceDBPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Calculate date range
//startDate := time.Now().Format("2006-01-02")
//endDate := time.Now().AddDate(0, 0, 5).Format("2006-01-02")

	startDate := time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	endDate := time.Date(2024, 4, 9, 0, 0, 0, 0, time.UTC).Format("2006-01-02")


  fmt.Printf("start: %s \n end: %s \n", startDate,endDate)

query := `
		SELECT weather_id, city_name, county_code, date, weather_type, temperature, weather_icon_url, google_weather_link
		FROM weather
		WHERE date >= ? AND date <= ?
	`
	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []WeatherRecord
	for rows.Next() {
		var record WeatherRecord
		err := rows.Scan(&record.WeatherID, &record.CityName, &record.CountryCode, &record.Date, &record.WeatherType, &record.Temperature, &record.WeatherIconURL, &record.GoogleWeatherLink)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// WeatherRecord holds weather data
type WeatherRecord struct {
	WeatherID        int
	CityName         string
	CountryCode      string
	Date             string
	WeatherType      string
	Temperature      float64
	WeatherIconURL   string
	GoogleWeatherLink string
}
