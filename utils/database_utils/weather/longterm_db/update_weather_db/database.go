
package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// AirportInfo represents the basic information of an airport necessary for weather data fetching
type AirportInfo struct {
	City    string
	Country string
}

// fetchAirports retrieves all airports with non-empty IATA codes from flights.db
func fetchAirports(db *sql.DB) ([]AirportInfo, error) {
	query := `SELECT city, country FROM airport_info WHERE iata IS NOT NULL AND iata != '';`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var airports []AirportInfo
	for rows.Next() {
		var ai AirportInfo
		if err := rows.Scan(&ai.City, &ai.Country); err != nil {
			return nil, err
		}
		airports = append(airports, ai)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return airports, nil
}

// initWeatherDB creates the weather database and table if it doesn't exist
func initWeatherDB(dbPath string) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Error opening weather.db: %v", err)
	}
	defer db.Close()

	createTableSQL := `CREATE TABLE IF NOT EXISTS Weather (
		WeatherID INTEGER PRIMARY KEY AUTOINCREMENT,
		CityName TEXT NOT NULL,
		CountryCode TEXT NOT NULL,
		Date TEXT NOT NULL,
		WeatherType TEXT NOT NULL,
		Temperature REAL NOT NULL,
		WeatherIconURL TEXT NOT NULL,
		GoogleWeatherLink TEXT NOT NULL
	);`

	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatalf("Error creating Weather table: %v", err)
	}
}
