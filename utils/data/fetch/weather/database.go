
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
  IATA    string
}

// fetchAirports retrieves all airports with non-empty IATA codes from flights.db
func fetchAirports(db *sql.DB) ([]AirportInfo, error) {

query := `SELECT a.city, a.country, a.iata
FROM airport a
JOIN city c ON LOWER(TRIM(a.city)) = LOWER(TRIM(c.city_ascii)) 
            AND LOWER(TRIM(a.country)) = LOWER(TRIM(c.iso2))  -- Using iso2 for country code
WHERE a.iata IS NOT NULL
AND a.iata != ''
AND a.city IS NOT NULL
AND a.country IS NOT NULL
AND c.include_tf = 1;
`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var airports []AirportInfo
	for rows.Next() {
		var ai AirportInfo
		if err := rows.Scan(&ai.City, &ai.Country, &ai.IATA); err != nil {
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

	createTableSQL := `CREATE TABLE IF NOT EXISTS all_weather (
		weather_id INTEGER PRIMARY KEY AUTOINCREMENT,
		city_name TEXT NOT NULL,
		country_code TEXT NOT NULL,
    iata TEXT NOT NULL,
		date TEXT NOT NULL,
		weather_type TEXT NOT NULL,
		temperature REAL NOT NULL,
		weather_icon_url TEXT NOT NULL,
		google_weather_link TEXT NOT NULL,
    wind_speed REAL NOT NULL,
    wpi FLOAT(10,1)
	);`

	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatalf("Error creating Weather table: %v", err)
	}
}
