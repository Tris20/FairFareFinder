
package main

import (
    "database/sql"
    "fmt"
    _ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
    "github.com/schollz/progressbar/v3"
)

type City struct {
    City       string
    IncludeTF  int
    CityAscii  string
    Lat        float64
    Lon        float64
    Country    string
    Iso2       string
    Iso3       string
    AdminName  sql.NullString
    Capital    sql.NullString
    Population sql.NullInt64
    Id         int
    IATACodes  []string
}

func main() {
    // Open locations.db
    db, err := sql.Open("sqlite3", "../../../../../../data/raw/locations/locations.db")
    if err != nil {
        fmt.Println(err)
        return
    }
    defer db.Close()

    // Query cities where include_tf == 1
    rows, err := db.Query("SELECT city, include_tf, city_ascii, lat, lon, country, iso2, iso3, admin_name, capital, population, id FROM city WHERE include_tf = 1")
    if err != nil {
        fmt.Println(err)
        return
    }
    defer rows.Close()

    var cities []City

    // Iterate over the rows
    for rows.Next() {
        var c City
        err = rows.Scan(&c.City, &c.IncludeTF, &c.CityAscii, &c.Lat, &c.Lon, &c.Country, &c.Iso2, &c.Iso3, &c.AdminName, &c.Capital, &c.Population, &c.Id)
        if err != nil {
            fmt.Println(err)
            return
        }
// Fetch IATA codes from the "airport" table
airports, err := db.Query("SELECT iata FROM airport WHERE LOWER(city) = LOWER(?) AND LOWER(country) = LOWER(?)", c.CityAscii, c.Iso2)
if err != nil {
    fmt.Println(err)
    continue
}
        
for airports.Next() {
    var iata string
    if err := airports.Scan(&iata); err != nil {
        fmt.Println("Error scanning IATA code:", err)
        continue
    }
    fmt.Printf("Fetched IATA: %s for city: %s\n", iata, c.CityAscii) // Debug output
    c.IATACodes = append(c.IATACodes, iata)
}
if len(c.IATACodes) == 0 {
    fmt.Printf("No IATA codes found for city: %s %s\n, ", c.CityAscii, c.Country) // Debug output
}
        airports.Close()

        cities = append(cities, c)
    }

    // Close and open new database
    db.Close()
    db, err = sql.Open("sqlite3", "../../../../../../data/compiled/new_main.db")
    if err != nil {
        fmt.Println(err)
        return
    }
    defer db.Close()

    // Initialize the progress bar
    bar := progressbar.NewOptions(len(cities),
        progressbar.OptionSetDescription("Inserting into new_main.db"),
        progressbar.OptionFullWidth(),
        progressbar.OptionSetPredictTime(false),
        progressbar.OptionShowCount(),
        progressbar.OptionShowIts(),
        progressbar.OptionSetItsString("cities"),
    )

    // Insert into the new database
    for _, c := range cities {
        query := "INSERT INTO location (city, country, iata_1, iata_2, iata_3, iata_4, iata_5, iata_6, iata_7, avg_wpi) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
        args := fillIATAs(c.CityAscii, c.Iso2, c.IATACodes)
        if _, err := db.Exec(query, args...); err != nil {
            fmt.Println(err)
            continue
        }
        bar.Add(1) // Increment the progress bar for each city processed
    }
    bar.Finish() // End the progress bar when loop is complete
}

func fillIATAs(city, country string, codes []string) []interface{} {
    result := make([]interface{}, 10) // Total 10 placeholders: city, country, 7 iatas, avg_wpi
    result[0] = city
    result[1] = country
    for i, code := range codes {
        if i >= 7 {
            break
        }
        result[i+2] = code
    }
    for i := len(codes) + 2; i < 9; i++ {
        result[i] = nil
    }
    result[9] = nil // avg_wpi as nil, replace or remove as needed
    return result
}

