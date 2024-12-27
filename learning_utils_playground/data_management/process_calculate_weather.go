package data_management

// main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Tris20/FairFareFinder/learning_utils_playground/config_handlers"
	db_manager "github.com/Tris20/FairFareFinder/learning_utils_playground/database_manager"
	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

type WeatherPleasantnessConfig struct {
	Conditions map[string]float64 `yaml:"conditions"`
}

type WeatherEntry struct {
	City              string
	Country           string
	IATA              string
	Date              string
	WeatherType       string
	Temperature       float64
	WindSpeed         float64
	WPI               float64
	WeatherIconURL    string
	GoogleWeatherLink string
}

func ProcessCalculateWeather() {
	if len(os.Args) == 4 {
		processCommandLineArguments()
	} else {
		weatherDBPath := "./testdata/weather.db"
		weatherPleasantnessYamlPath := "../config/weatherPleasantness.yaml"
		ProcessDatabaseEntries(weatherDBPath, weatherPleasantnessYamlPath)
	}
}

func processCommandLineArguments() {
	temp, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		log.Fatal("Invalid temperature input:", err)
	}

	windSpeed, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		log.Fatal("Invalid wind speed input:", err)
	}

	condition := os.Args[3]

	config, err := config_handlers.LoadWeatherPleasantnessConfig("../../../../../config/weatherPleasantness.yaml")
	if err != nil {
		log.Fatal("Error loading weather pleasantness config:", err)
	}

	wpi := weatherPleasantness(temp, windSpeed, condition, config)
	fmt.Printf("Weather Pleasantness Index: %.2f\n", wpi)
}

func collectWeatherEntries(db *sql.DB) ([]WeatherEntry, error) {
	rows, err := db.Query(`
	SELECT city_name, country_code, iata, date, weather_type, temperature, wind_speed, weather_icon_url, google_weather_link
	FROM all_weather
	WHERE datetime(date) > datetime('now', 'localtime') AND city_name != ''
`)
	if err != nil {
		log.Fatal("Error querying database:", err)
	}
	defer rows.Close()

	var entries []WeatherEntry
	for rows.Next() {
		var entry WeatherEntry
		if err := rows.Scan(&entry.City, &entry.Country, &entry.IATA, &entry.Date, &entry.WeatherType, &entry.Temperature, &entry.WindSpeed, &entry.WeatherIconURL, &entry.GoogleWeatherLink); err != nil {
			log.Fatal("Error scanning database row:", err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func insertCurrentWeatherEntries(db *sql.DB, entries []WeatherEntry,
	config config_handlers.WeatherPleasantnessConfig) error {
	bar := progressbar.Default(int64(len(entries)))

	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Error starting transaction:", err)
	}

	stmt, err := tx.Prepare(`INSERT INTO current_weather (city_name, country_code, iata, date, weather_type, temperature, wind_speed, wpi, weather_icon_url, google_weather_link) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		log.Fatal("Failed to prepare SQL statement:", err)
	}
	defer stmt.Close()

	for _, entry := range entries {
		entry.WPI = weatherPleasantness(entry.Temperature, entry.WindSpeed, entry.WeatherType, config)
		_, err = stmt.Exec(entry.City, entry.Country, entry.IATA, entry.Date, entry.WeatherType, entry.Temperature, entry.WindSpeed, entry.WPI, entry.WeatherIconURL, entry.GoogleWeatherLink)
		if err != nil {
			tx.Rollback()
			log.Fatal("Failed to insert data into current_weather:", err)
		}
		bar.Add(1)
	}
	bar.Finish()

	if err := tx.Commit(); err != nil {
		log.Fatal("Error committing transaction:", err)
	}
	return nil
}

func ProcessDatabaseEntries(weatherDBPath, weatherPleasantnessYamlPath string) error {
	db, err := sql.Open("sqlite3", weatherDBPath)
	if err != nil {
		log.Print("Error opening database:", err)
		return err
	}
	defer db.Close()

	// make sure tables are setup
	err = db_manager.RecreateTable(db, &db_manager.CurrentWeather{})
	if err != nil {
		log.Print("Failed to recreate current_weather table:", err)
		return err
	}

	err = db_manager.CreateTable(db, &db_manager.AllWeather{})
	if err != nil {
		log.Print("Failed to create all_weather table:", err)
		return err
	}

	// collect weather entries
	entries, err := collectWeatherEntries(db)
	if err != nil {
		log.Print("Error collecting weather entries:", err)
		return err
	}

	config, err := config_handlers.LoadWeatherPleasantnessConfig(weatherPleasantnessYamlPath)
	if err != nil {
		log.Fatal("Error loading weather pleasantness config:", err)
	}

	return insertCurrentWeatherEntries(db, entries, config)
}

func weatherPleasantness(temp float64, wind float64, cond string, config config_handlers.WeatherPleasantnessConfig) float64 {
	weightTemp := 5.0
	weightWind := 1.0
	weightCond := 2.0

	tempIndex := tempPleasantness(temp) * weightTemp
	windIndex := windPleasantness(wind) * weightWind
	weatherIndex := weatherCondPleasantness(cond, config) * weightCond

	return (tempIndex + windIndex + weatherIndex) / (weightTemp + weightWind + weightCond)
}

// utils
// Simple math functions and checks go here

func interpolate(temp, temp1, temp2, score1, score2 float64) float64 {
	return ((temp-temp1)/(temp2-temp1))*(score2-score1) + score1
}

// tempPleasantness, windPleasantness, weatherCondPleasantness defined here.

func tempPleasantness(temperature float64) float64 {
	// Optimal range
	if temperature >= 22 && temperature <= 26 {
		return 10
	}

	//     Interpolation between key temperatures below the optimal range
	if temperature > 18 && temperature < 22 {
		return interpolate(temperature, 18, 22, 7, 10)
	}

	// Interpolation between key temperatures above the optimal range
	if temperature > 26 && temperature < 40 {
		return interpolate(temperature, 26, 40, 10, 0)
	}

	// Below 18 down to 0
	if temperature >= 5 && temperature <= 18 {
		return interpolate(temperature, 5, 18, 0, 7)
	}

	// Anything below 0 or above 40
	if temperature <= 5 || temperature >= 40 {
		return 0
	}

	return 0 // Default case if needed
}

// windPleasantness returns a value between 0 and 10 for wind condition pleasantness
func windPleasantness(windSpeed float64) float64 {
	worstWind := 13.8
	if windSpeed >= worstWind {
		return 0
	} else {
		return 10 - windSpeed*10/worstWind
	}
}

// weatherCondPleasantness returns a value between 0 and 10 for weather condition pleasantness
func weatherCondPleasantness(cond string, config config_handlers.WeatherPleasantnessConfig) float64 {
	pleasantness, ok := config.Conditions[cond]
	if !ok {
		return 0
	}
	return pleasantness
}
