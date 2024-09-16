
package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

type WeatherData struct {
	CityName          string
	CountryCode       string
	Date              string
	Temperature       float64
	WPI               float64
	WeatherIconURL    string
	GoogleWeatherLink string
}

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
	db, err := sql.Open("sqlite3", "../../../../../../data/raw/weather/weather.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := `
	SELECT city_name, country_code, date, AVG(temperature) AS avg_temp, AVG(wpi) AS avg_wpi, weather_icon_url, google_weather_link
	FROM current_weather
	WHERE strftime('%H:%M:%S', date) IN ('11:00:00', '14:00:00', '17:00:00')
	GROUP BY city_name, country_code, strftime('%Y-%m-%d', date)
	`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var weathers []CompiledWeather
	for rows.Next() {
		var wd WeatherData
		err := rows.Scan(&wd.CityName, &wd.CountryCode, &wd.Date, &wd.Temperature, &wd.WPI, &wd.WeatherIconURL, &wd.GoogleWeatherLink)
		if err != nil {
			log.Fatal(err)
		}
		formattedTemp, _ := strconv.ParseFloat(fmt.Sprintf("%.1f", wd.Temperature), 64)
		formattedWPI, _ := strconv.ParseFloat(fmt.Sprintf("%.1f", wd.WPI), 64)
		weathers = append(weathers, CompiledWeather{
			City:           wd.CityName,
			Country:        wd.CountryCode,
			Date:           strings.Split(wd.Date, " ")[0],
			AvgDaytimeTemp: formattedTemp,
			WeatherIcon:    wd.WeatherIconURL,
			GoogleURL:      wd.GoogleWeatherLink,
			AvgDaytimeWPI:  formattedWPI,
		})
	}

	compiledDB, err := sql.Open("sqlite3", "../../../../../../data/compiled/new_main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer compiledDB.Close()

	stmt, err := compiledDB.Prepare("INSERT INTO weather (city, country, date, avg_daytime_temp, weather_icon, google_url, avg_daytime_wpi) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	bar := progressbar.Default(int64(len(weathers)))
	for _, w := range weathers {
		_, err := stmt.Exec(w.City, w.Country, w.Date, w.AvgDaytimeTemp, w.WeatherIcon, w.GoogleURL, w.AvgDaytimeWPI)
		if err != nil {
			log.Fatal(err)
		}
		bar.Add(1)
	}
	fmt.Println("Data successfully transferred to new_main.db")
}

