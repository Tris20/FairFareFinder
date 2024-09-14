
package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

// CurrentWeather represents the structure of the current_weather table
type CurrentWeather struct {
	CityName          string
	CountryCode       string
	IATA              string
	Date              string
	WeatherType       string
	Temperature       float64
	WeatherIconURL    string
	GoogleWeatherLink string
	WindSpeed         float64
	WPI               float64
}

// CompiledWeather represents the structure of the weather table for later use
type CompiledWeather struct {
	City           string
	Country        string
	Date           string
	AvgDaytimeTemp float64
	WeatherIcon    string
	GoogleURL      string
	AvgDaytimeWPI  float64
}

func main() {
	// Open the current weather database
	db, err := sql.Open("sqlite3", "../../../../../../data/raw/weather/weather.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Query the current_weather table
	rows, err := db.Query("SELECT city_name, country_code, date, temperature, weather_icon_url, google_weather_link, wpi FROM current_weather")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var weathers []CompiledWeather
	for rows.Next() {
		var cw CurrentWeather
		if err := rows.Scan(&cw.CityName, &cw.CountryCode, &cw.Date, &cw.Temperature, &cw.WeatherIconURL, &cw.GoogleWeatherLink, &cw.WPI); err != nil {
			log.Fatal(err)
		}
		t, _ := time.Parse("2006-01-02 15:04:05", cw.Date)
		hour := t.Hour()
		if hour == 23 || hour == 2 || hour == 5 {
			continue // skip these entries
		}
		if hour == 14 {
			weathers = append(weathers, CompiledWeather{
				City:           cw.CityName,
				Country:        cw.CountryCode,
				Date:           strings.Split(cw.Date, " ")[0],
				AvgDaytimeTemp: cw.Temperature,
				WeatherIcon:    cw.WeatherIconURL,
				GoogleURL:      cw.GoogleWeatherLink,
				AvgDaytimeWPI:  cw.WPI,
			})
		}
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// Open the compiled weather database
	compiledDB, err := sql.Open("sqlite3", "../../../../../../data/compiled/new_main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer compiledDB.Close()

	// Prepare the insert statement
	stmt, err := compiledDB.Prepare("INSERT INTO weather (city, country, date, avg_daytime_temp, weather_icon, google_url, avg_daytime_wpi) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Create a new progress bar
	bar := progressbar.Default(int64(len(weathers)))

	// Insert compiled weather data into the new database and update progress bar
	for _, w := range weathers {
		_, err := stmt.Exec(w.City, w.Country, w.Date, w.AvgDaytimeTemp, w.WeatherIcon, w.GoogleURL, w.AvgDaytimeWPI)
		if err != nil {
			log.Fatal(err)
		}
		bar.Add(1)
	}
	fmt.Println("Data successfully transferred to new_main.db")
}

