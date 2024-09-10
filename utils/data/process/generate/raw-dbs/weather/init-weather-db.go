package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open the SQLite database.
	db, err := sql.Open("sqlite3", "../../../../../../data/raw/weather/weather.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the 'weather' table.
	createTableSQL := `CREATE TABLE IF NOT EXISTS weather (
		weather_id INTEGER PRIMARY KEY AUTOINCREMENT,
		city_name TEXT NOT NULL,
		country_code TEXT NOT NULL,
		iata TEXT NOT NULL,
		date TEXT NOT NULL,
		weather_type TEXT NOT NULL,
		temperature REAL NOT NULL,
		weather_icon_url TEXT NOT NULL,
		google_weather_link TEXT NOT NULL,
		wind_speed REAL NOT NULL
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Weather table created successfully.")
}

