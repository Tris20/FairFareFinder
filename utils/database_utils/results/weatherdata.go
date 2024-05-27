
package main

import (
	"database/sql"
	"fmt"
	"time"

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
	startDate := time.Date(2024, 5, 25, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	endDate := time.Date(2024, 5, 30, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	fmt.Printf("start: %s \n end: %s \n", startDate, endDate)

	query := `
		SELECT weather_id, city_name, country_code, iata, date, weather_type, temperature, weather_icon_url, google_weather_link, wind_speed
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
		var windSpeed sql.NullFloat64
		err := rows.Scan(&record.WeatherID, &record.CityName, &record.CountryCode, &record.IATA, &record.Date, &record.WeatherType, &record.Temperature, &record.WeatherIconURL, &record.GoogleWeatherLink, &windSpeed)
		if err != nil {
			return nil, err
		}
		if windSpeed.Valid {
			record.WindSpeed = windSpeed.Float64
		} else {
			record.WindSpeed = 0.0 // or any default value you prefer
		}
		records = append(records, record)
	}

	return records, nil
}


