package main

import (
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open the raw flights database.
	rawDBPath := filepath.Join("../../../data/raw/flights", "flights.db")
	rawDB, err := sql.Open("sqlite3", rawDBPath)
	if err != nil {
		log.Fatalf("Failed to open raw flights database: %v", err)
	}
	defer rawDB.Close()

	// Open the generated flight-prices database.
	fpDBPath := filepath.Join("../../../data/generated", "flight-prices.db")
	fpDB, err := sql.Open("sqlite3", fpDBPath)
	if err != nil {
		log.Fatalf("Failed to open flight-prices database: %v", err)
	}
	defer fpDB.Close()

	// Open the flight price modifiers database.
	modDBPath := filepath.Join("../../../data/generated", "flight_price_modifiers.db")
	modDB, err := sql.Open("sqlite3", modDBPath)
	if err != nil {
		log.Fatalf("Failed to open flight_price_modifiers database: %v", err)
	}
	defer modDB.Close()

	// Fix: Open the raw locations database using the corrected path.
	locDBPath := filepath.Join("../../../data/raw/locations", "locations.db")
	locDB, err := sql.Open("sqlite3", locDBPath)
	if err != nil {
		log.Fatalf("Failed to open locations database: %v", err)
	}
	defer locDB.Close()

	// Query all unique routes from the schedule table.
	uniqueRoutesQuery := `
		SELECT DISTINCT departureAirport, arrivalAirport
		FROM schedule
		ORDER BY departureAirport, arrivalAirport;
	`
	rows, err := rawDB.Query(uniqueRoutesQuery)
	if err != nil {
		log.Fatalf("Failed to query unique routes: %v", err)
	}
	defer rows.Close()

	// Prepare an INSERT statement for the routes table.
	insertStmt, err := fpDB.Prepare(`
		INSERT INTO routes (
			origin_city_name, origin_country, origin_iata, origin_population,
			destination_city_name, destination_country, destination_iata, destination_population,
			route_frequency, most_common_airline, most_common_aircraft, most_common_aircraft_seating_capacity,
			route_classification
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`)
	if err != nil {
		log.Fatalf("Failed to prepare insert statement: %v", err)
	}
	defer insertStmt.Close()

	// Process each unique route.
	for rows.Next() {
		var departureAirport, arrivalAirport string
		if err := rows.Scan(&departureAirport, &arrivalAirport); err != nil {
			log.Printf("Failed to scan route: %v", err)
			continue
		}

		// 1. Compute route frequency.
		var routeFrequency int
		err = rawDB.QueryRow(
			`SELECT COUNT(*) FROM schedule WHERE departureAirport = ? AND arrivalAirport = ?;`,
			departureAirport, arrivalAirport,
		).Scan(&routeFrequency)
		if err != nil {
			log.Printf("Failed to count flights for route %s -> %s: %v", departureAirport, arrivalAirport, err)
			continue
		}

		// 2. Determine the most common airline.
		var mostCommonAirline string
		err = rawDB.QueryRow(
			`SELECT airline_name FROM schedule 
			 WHERE departureAirport = ? AND arrivalAirport = ? AND airline_name IS NOT NULL
			 GROUP BY airline_name
			 ORDER BY COUNT(*) DESC LIMIT 1;`,
			departureAirport, arrivalAirport,
		).Scan(&mostCommonAirline)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Failed to get most common airline for route %s -> %s: %v", departureAirport, arrivalAirport, err)
			continue
		}

		// 3. Determine the most common aircraft.
		var mostCommonAircraft string
		err = rawDB.QueryRow(
			`SELECT aircraft_model FROM schedule 
			 WHERE departureAirport = ? AND arrivalAirport = ? AND aircraft_model IS NOT NULL
			 GROUP BY aircraft_model
			 ORDER BY COUNT(*) DESC LIMIT 1;`,
			departureAirport, arrivalAirport,
		).Scan(&mostCommonAircraft)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Failed to get most common aircraft for route %s -> %s: %v", departureAirport, arrivalAirport, err)
			continue
		}

		// 4. Look up seating capacity for the most common aircraft from the modifiers database.
		var seatingCapacity int
		err = modDB.QueryRow(
			`SELECT seating_capacity FROM aircraft_capacity_lookup 
			 WHERE aircraft_model = ? LIMIT 1;`,
			mostCommonAircraft,
		).Scan(&seatingCapacity)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("No seating capacity found for aircraft %s on route %s -> %s", mostCommonAircraft, departureAirport, arrivalAirport)
				seatingCapacity = 0
			} else {
				log.Printf("Failed to get seating capacity for aircraft %s: %v", mostCommonAircraft, err)
				seatingCapacity = 0
			}
		}

		// 5. Look up airport details for origin.
		var originCity, originCountry string
		err = locDB.QueryRow(
			`SELECT city, country FROM airport WHERE iata = ? LIMIT 1;`,
			departureAirport,
		).Scan(&originCity, &originCountry)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("No airport info found for origin iata %s", departureAirport)
				originCity, originCountry = "", ""
			} else {
				log.Printf("Failed to get airport info for origin iata %s: %v", departureAirport, err)
				originCity, originCountry = "", ""
			}
		}

		// 6. Look up population for origin city.
		var originPopulation int
		err = locDB.QueryRow(
			`SELECT population FROM city WHERE city_ascii = ? AND iso2 = ? LIMIT 1;`,
			originCity, originCountry,
		).Scan(&originPopulation)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("No population info found for origin city %s, country %s", originCity, originCountry)
				originPopulation = 0
			} else {
				log.Printf("Failed to get population for origin city %s, country %s: %v", originCity, originCountry, err)
				originPopulation = 0
			}
		}

		// 7. Look up airport details for destination.
		var destCity, destCountry string
		err = locDB.QueryRow(
			`SELECT city, country FROM airport WHERE iata = ? LIMIT 1;`,
			arrivalAirport,
		).Scan(&destCity, &destCountry)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("No airport info found for destination iata %s", arrivalAirport)
				destCity, destCountry = "", ""
			} else {
				log.Printf("Failed to get airport info for destination iata %s: %v", arrivalAirport, err)
				destCity, destCountry = "", ""
			}
		}

		// 8. Look up population for destination city.
		var destPopulation int
		err = locDB.QueryRow(
			`SELECT population FROM city WHERE city_ascii = ? AND iso2 = ? LIMIT 1;`,
			destCity, destCountry,
		).Scan(&destPopulation)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("No population info found for destination city %s, country %s", destCity, destCountry)
				destPopulation = 0
			} else {
				log.Printf("Failed to get population for destination city %s, country %s: %v", destCity, destCountry, err)
				destPopulation = 0
			}
		}

		// 9. Look up route classification from the modifiers database.
		var routeClassification string
		err = modDB.QueryRow(
			`SELECT route_classification FROM route_classification_lookup 
			 WHERE departureAirport = ? AND arrivalAirport = ? LIMIT 1;`,
			departureAirport, arrivalAirport,
		).Scan(&routeClassification)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("No route classification found for route %s -> %s", departureAirport, arrivalAirport)
				routeClassification = ""
			} else {
				log.Printf("Failed to get route classification for route %s -> %s: %v", departureAirport, arrivalAirport, err)
				routeClassification = ""
			}
		}

		// Insert the aggregated data into the routes table.
		_, err = insertStmt.Exec(
			originCity,          // origin_city_name
			originCountry,       // origin_country
			departureAirport,    // origin_iata
			originPopulation,    // origin_population
			destCity,            // destination_city_name
			destCountry,         // destination_country
			arrivalAirport,      // destination_iata
			destPopulation,      // destination_population
			routeFrequency,      // route_frequency
			mostCommonAirline,   // most_common_airline
			mostCommonAircraft,  // most_common_aircraft
			seatingCapacity,     // most_common_aircraft_seating_capacity
			routeClassification, // route_classification
		)
		if err != nil {
			log.Printf("Failed to insert route %s -> %s: %v", departureAirport, arrivalAirport, err)
			continue
		}

		log.Printf("Inserted route %s -> %s: frequency=%d, origin=%s (%s, pop=%d), destination=%s (%s, pop=%d), aircraft=%s (seats=%d), classification=%s",
			departureAirport, arrivalAirport, routeFrequency,
			originCity, originCountry, originPopulation,
			destCity, destCountry, destPopulation,
			mostCommonAircraft, seatingCapacity, routeClassification)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error reading unique routes: %v", err)
	}

	fmt.Println("Successfully processed and inserted unique routes into flight-prices.db")

	// Call the "flight-duration" executable with the argument "calculate_prices".
	// The executable is located at "/utils/data/process/calculate/flights/flight-duration".
	// Set the working directory for the flight-duration executable.
	flightDurationDir := filepath.Join("../../../utils/data/process/calculate/flights/flight-duration")
	// Use a relative path to the executable since we are setting the working directory.
	cmd := exec.Command("./flight-duration", "calculate_prices")
	// Set the command's working directory to the flight-duration directory.
	cmd.Dir = flightDurationDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to execute flight-duration: %v\nOutput: %s", err, string(output))
	}
	fmt.Printf("flight-duration output:\n%s\n", string(output))
}
