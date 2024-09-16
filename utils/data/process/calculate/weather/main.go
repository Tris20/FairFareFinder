
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "strconv"

    "github.com/schollz/progressbar/v3"
    _ "github.com/mattn/go-sqlite3"
    "github.com/Tris20/FairFareFinder/config/handlers"
)

type WeatherPleasantnessConfig struct {
    Conditions map[string]float64 `yaml:"conditions"`
}

type WeatherEntry struct {
    City             string
    Country          string
    IATA             string
    Date             string
    WeatherType      string
    Temperature      float64
    WindSpeed        float64
    WPI              float64
    WeatherIconURL   string
    GoogleWeatherLink string
}

func main() {
    if len(os.Args) == 4 {
        processCommandLineArguments()
    } else {
        processDatabaseEntries()
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

func processDatabaseEntries() {
    db, err := sql.Open("sqlite3", "../../../../../data/raw/weather/weather.db")
    if err != nil {
        log.Fatal("Error opening database:", err)
    }
    defer db.Close()

    // Drop the current_weather table if it exists
    _, err = db.Exec("DROP TABLE IF EXISTS current_weather")
    if err != nil {
        log.Fatal("Failed to drop current_weather table:", err)
    }

    // Create the current_weather table
    _, err = db.Exec(`
        CREATE TABLE current_weather (
            city_name TEXT,
            country_code TEXT,
            iata TEXT,
            date TEXT,
            weather_type TEXT,
            temperature REAL,
            weather_icon_url TEXT,
            google_weather_link TEXT,
            wind_speed REAL,
            wpi REAL
        )
    `)
    if err != nil {
        log.Fatal("Failed to create current_weather table:", err)
    }

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

    config, err := config_handlers.LoadWeatherPleasantnessConfig("../../../../../config/weatherPleasantness.yaml")
    if err != nil {
        log.Fatal("Error loading weather pleasantness config:", err)
    }

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

