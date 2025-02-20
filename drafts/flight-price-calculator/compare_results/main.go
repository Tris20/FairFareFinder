package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open the routes database (flight-prices.db)
	routesDB, err := sql.Open("sqlite3", "../../../data/generated/flight-prices.db")
	if err != nil {
		log.Fatal(err)
	}
	defer routesDB.Close()

	// Attach the flight database (main.db) as "flightdb"
	attachStmt := `ATTACH DATABASE '../../../data/compiled/main.db' AS flightdb;`
	_, err = routesDB.Exec(attachStmt)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new output database to store the merged data.
	outDB, err := sql.Open("sqlite3", "price-comparison.db")
	if err != nil {
		log.Fatal(err)
	}
	defer outDB.Close()

	// Create the new table with computed columns at the end:
	// error_multiple is second last and error_direction is last.
	createTableSQL := `
CREATE TABLE IF NOT EXISTS route_comparison (
	id INTEGER PRIMARY KEY,
	origin_city_name TEXT,
	origin_country TEXT,
	origin_iata TEXT,
	origin_population INTEGER,
	destination_city_name TEXT,
	destination_country TEXT,
	destination_iata TEXT,
	destination_population INTEGER,
	route_frequency INTEGER,
	route_classification TEXT,
	most_common_airline TEXT,
	most_common_aircraft TEXT,
	most_common_aircraft_seating_capacity INTEGER,
	duration_in_minutes INTEGER,
	duration_in_hours REAL,
	duration_in_hours_rounded INTEGER,
	duration_hour_dot_mins TEXT,
	calculated_price REAL,
	actual_price REAL,
	price_difference REAL,
	error_multiple REAL,
	error_direction TEXT
);
	`
	_, err = outDB.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Query to join the two tables, returning columns in the same order as the output table.
	query := `
SELECT
    r.id,
    r.origin_city_name,
    r.origin_country,
    r.origin_iata,
    r.origin_population,
    r.destination_city_name,
    r.destination_country,
    r.destination_iata,
    r.destination_population,
    r.route_frequency,
    r.route_classification,
    r.most_common_airline,
    r.most_common_aircraft,
    r.most_common_aircraft_seating_capacity,
    r.duration_in_minutes,
    r.duration_in_hours,
    r.duration_in_hours_rounded,
    r.duration_hour_dot_mins,
    r.calculated_price,
    f.price_this_week AS actual_price,
    (r.calculated_price - f.price_this_week) AS price_difference,
    ROUND(
      CASE
        WHEN f.price_this_week IS NULL OR f.price_this_week = 0 THEN 0
        WHEN r.calculated_price >= f.price_this_week THEN r.calculated_price / f.price_this_week
        ELSE f.price_this_week / r.calculated_price
      END, 2
    ) AS error_multiple,
    CASE
        WHEN f.price_this_week IS NULL OR f.price_this_week = 0 THEN 'unknown'
        WHEN r.calculated_price > f.price_this_week THEN 'too high'
        WHEN r.calculated_price < f.price_this_week THEN 'too low'
        ELSE 'equal'
    END AS error_direction
FROM routes r
JOIN flightdb.flight f
    ON r.origin_iata = f.origin_iata AND r.destination_iata = f.destination_iata;
`

	rows, err := routesDB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Prepare the insert statement for the output table.
	insertSQL := `
INSERT OR REPLACE INTO route_comparison (
    id, origin_city_name, origin_country, origin_iata, origin_population,
    destination_city_name, destination_country, destination_iata, destination_population,
    route_frequency, route_classification, most_common_airline, most_common_aircraft,
    most_common_aircraft_seating_capacity, duration_in_minutes, duration_in_hours,
    duration_in_hours_rounded, duration_hour_dot_mins, calculated_price,
    actual_price, price_difference, error_multiple, error_direction
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
`

	stmt, err := outDB.Prepare(insertSQL)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Iterate over the query results and insert them into the new table.
	for rows.Next() {
		var (
			id                                                           int
			originCityName, originCountry, originIATA                    string
			originPopulation                                             int
			destinationCityName, destinationCountry, destinationIATA     string
			destinationPopulation                                        int
			routeFrequency                                               int
			routeClassification, mostCommonAirline, mostCommonAircraft   string
			mostCommonAircraftSeatingCapacity                            int
			durationInMinutes                                            int
			durationInHours                                              float64
			durationInHoursRounded                                       int
			durationHourDotMins                                          string
			calculatedPrice, actualPrice, priceDifference, errorMultiple float64
			errorDirection                                               string
		)
		err = rows.Scan(
			&id,
			&originCityName,
			&originCountry,
			&originIATA,
			&originPopulation,
			&destinationCityName,
			&destinationCountry,
			&destinationIATA,
			&destinationPopulation,
			&routeFrequency,
			&routeClassification,
			&mostCommonAirline,
			&mostCommonAircraft,
			&mostCommonAircraftSeatingCapacity,
			&durationInMinutes,
			&durationInHours,
			&durationInHoursRounded,
			&durationHourDotMins,
			&calculatedPrice,
			&actualPrice,
			&priceDifference,
			&errorMultiple,
			&errorDirection,
		)
		if err != nil {
			log.Fatal(err)
		}

		_, err = stmt.Exec(
			id,
			originCityName,
			originCountry,
			originIATA,
			originPopulation,
			destinationCityName,
			destinationCountry,
			destinationIATA,
			destinationPopulation,
			routeFrequency,
			routeClassification,
			mostCommonAirline,
			mostCommonAircraft,
			mostCommonAircraftSeatingCapacity,
			durationInMinutes,
			durationInHours,
			durationInHoursRounded,
			durationHourDotMins,
			calculatedPrice,
			actualPrice,
			priceDifference,
			errorMultiple,
			errorDirection,
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Database creation complete. Output stored in price-comparison.db")
}
