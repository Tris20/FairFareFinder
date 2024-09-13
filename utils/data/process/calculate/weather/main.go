
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
    City        string
    Country     string
    IATA        string
    Date        string
    WeatherType string
    Temperature float64
    WindSpeed   float64
    WPI         float64
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

    rows, err := db.Query(`SELECT city_name, country_code, iata, date, weather_type, temperature, wind_speed FROM weather WHERE datetime(date) > datetime('now', 'localtime')`)
    if err != nil {
        log.Fatal("Error querying database:", err)
    }
    defer rows.Close()

    var entries []WeatherEntry
    for rows.Next() {
        var entry WeatherEntry
        if err := rows.Scan(&entry.City, &entry.Country, &entry.IATA, &entry.Date, &entry.WeatherType, &entry.Temperature, &entry.WindSpeed); err != nil {
            log.Fatal("Error scanning database row:", err)
        }
        entries = append(entries, entry)
    }

    config, err := config_handlers.LoadWeatherPleasantnessConfig("../../../../../config/weatherPleasantness.yaml")
    if err != nil {
        log.Fatal("Error loading weather pleasantness config:", err)
    }


    for i := range entries {
        entries[i].WPI = weatherPleasantness(entries[i].Temperature, entries[i].WindSpeed, entries[i].WeatherType, config)
        //bar.Add(1)
    }
    
   // Initialize progress bar for processing entries
    bar := progressbar.Default(int64(len(entries)))

    // Start transaction
    tx, err := db.Begin()
    if err != nil {
        log.Fatal("Error starting transaction:", err)
    }

    // Prepare the SQL statement once
    stmt, err := tx.Prepare(`UPDATE weather SET wpi = ? WHERE city_name = ? AND country_code = ? AND iata = ? AND date = ?`)
    if err != nil {
        tx.Rollback()
        log.Fatal("Failed to prepare SQL statement:", err)
    }
    defer stmt.Close()

    // Execute updates in a batch
    for i, entry := range entries {
        _, err = stmt.Exec(entry.WPI, entry.City, entry.Country, entry.IATA, entry.Date)
        if err != nil {
            tx.Rollback()
            log.Fatal("Failed to update database:", err)
        }
        // Update progress bar less frequently to improve performance
        if (i+1) % 100 == 0 || i == len(entries)-1 { // Also ensure to update progress on the last item
            bar.Add(100)
        }
    }
    bar.Finish()

    // Commit transaction
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


