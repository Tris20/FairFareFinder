package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
	"log"
	"math"
	"os"
)

// Haversine formula to calculate distance between two coordinates
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0 // Earth radius in kilometers
	lat1, lon1, lat2, lon2 = degreesToRadians(lat1), degreesToRadians(lon1), degreesToRadians(lat2), degreesToRadians(lon2)
	dlat := lat2 - lat1
	dlon := lon2 - lon1

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

// Convert degrees to radians
func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

// Calculate estimated flight time based on distance and average speed
func calculateAdjustedFlightTime(distance float64) float64 {
	const averageSpeed = 900.0        // Average flight speed in km/h
	const takeoffLandingBuffer = 0.40 // Buffer for takeoff/landing in hours (30 minutes)
	const routeMultiplier = 1.10      // Multiplier for non-direct routes

	adjustedDistance := distance * routeMultiplier
	cruiseTime := adjustedDistance / averageSpeed
	totalFlightTime := cruiseTime + takeoffLandingBuffer
	return totalFlightTime
}

// Convert flight duration into hours.minutes and rounded hours
func formatDuration(hours float64) (float64, int, float64) {
	// Total minutes
	totalMinutes := int(math.Round(hours * 60))
	hoursPart := totalMinutes / 60
	minutesPart := totalMinutes % 60

	// Round minutes to the nearest valid value (0, 20, 40)
	switch {
	case minutesPart < 10:
		minutesPart = 0
	case minutesPart < 30:
		minutesPart = 20
	default:
		minutesPart = 40
	}

	// Calculate hours.minutes format
	durationHourDotMins := float64(hoursPart) + float64(minutesPart)/100

	// Calculate rounded hours
	durationHoursRounded := float64(hoursPart)
	if minutesPart == 40 {
		durationHoursRounded += 1.0
	}

	return durationHourDotMins, minutesPart, durationHoursRounded
}

func main() {
	// Determine mode from arguments.
	// Default mode: update flight durations in the main DB (flight table)
	// "calculate_prices" mode: update flight durations in flight-prices.db (routes table)
	mode := "default"
	if len(os.Args) > 1 && os.Args[1] == "calculate_prices" {
		mode = "calculate_prices"
	}

	var mainDBPath string
	var tableName string

	if mode == "calculate_prices" {
		mainDBPath = "../../../../../../data/generated/flight-prices.db"
		tableName = "routes"
		log.Printf("Running in calculate_prices mode. Updating %s table in %s", tableName, mainDBPath)
	} else {
		mainDBPath = "../../../../../../data/compiled/new_main.db"
		tableName = "flight"
		log.Printf("Running in default mode. Updating %s table in %s", tableName, mainDBPath)
	}

	// Open the main database.
	mainDB, err := sql.Open("sqlite3", mainDBPath)
	if err != nil {
		log.Fatalf("Failed to open main database: %v", err)
	}
	defer mainDB.Close()

	// The locations database (for airport coordinates) remains the same.
	locationsDBPath := "../../../../../../data/raw/locations/locations.db"
	log.Printf("Connecting to locations database at: %s", locationsDBPath)
	locationsDB, err := sql.Open("sqlite3", locationsDBPath)
	if err != nil {
		log.Fatalf("Failed to open locations database: %v", err)
	}
	defer locationsDB.Close()

	// Depending on the mode, query from the appropriate table.
	query := fmt.Sprintf(`SELECT id, origin_iata, destination_iata FROM %s`, tableName)
	log.Printf("Querying routes from table: %s", tableName)
	rows, err := mainDB.Query(query)
	if err != nil {
		log.Fatalf("Failed to query %s table: %v", tableName, err)
	}
	defer rows.Close()

	// Begin transaction for updates.
	tx, err := mainDB.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			log.Printf("Rolling back transaction due to error: %v", err)
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				log.Fatalf("Failed to commit transaction: %v", err)
			}
			log.Printf("Transaction committed successfully")
		}
	}()

	// Initialize progress bar.
	bar := progressbar.NewOptions(-1, progressbar.OptionSetDescription("Calculating flight durations"))
	var totalRoutes int

	// Iterate through each route.
	for rows.Next() {
		totalRoutes++
		var id int
		var originIATA, destinationIATA string
		if err := rows.Scan(&id, &originIATA, &destinationIATA); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}

		log.Printf("Processing route ID %d: %s -> %s", id, originIATA, destinationIATA)

		// Get coordinates for origin and destination airports.
		originLat, originLon, err := getAirportCoordinates(locationsDB, originIATA)
		if err != nil {
			log.Printf("Skipping route ID %d due to missing origin coordinates: %v", id, err)
			bar.Add(1)
			continue
		}

		destLat, destLon, err := getAirportCoordinates(locationsDB, destinationIATA)
		if err != nil {
			log.Printf("Skipping route ID %d due to missing destination coordinates: %v", id, err)
			bar.Add(1)
			continue
		}

		// Calculate distance and flight time.
		distance := haversine(originLat, originLon, destLat, destLon)
		log.Printf("Calculated distance for route ID %d: %.2f km", id, distance)
		flightTime := calculateAdjustedFlightTime(distance)
		log.Printf("Route ID %d: Calculated flightTime = %.2f hours", id, flightTime)

		// Format the duration values.
		durationHourDotMins, durationMinutes, durationHoursRounded := formatDuration(flightTime)
		hoursPart := int(durationHourDotMins) // integer hours part

		log.Printf("Route ID %d: duration_hour_dot_mins = %.2f, duration_in_minutes = %d, duration_in_hours = %d, duration_in_hours_rounded = %.2f",
			id, durationHourDotMins, durationMinutes, hoursPart, durationHoursRounded)

		// Build the update query using the same duration column names.
		updateQuery := fmt.Sprintf(`
			UPDATE %s
			SET duration_hour_dot_mins = ?,
			    duration_in_minutes = ?,
			    duration_in_hours = ?,
			    duration_in_hours_rounded = ?
			WHERE id = ?`, tableName)

		_, err = tx.Exec(updateQuery, durationHourDotMins, durationMinutes, float64(hoursPart), durationHoursRounded, id)
		if err != nil {
			log.Printf("Failed to update duration for route ID %d: %v", id, err)
		} else {
			log.Printf("Successfully updated route ID %d", id)
		}

		bar.Add(1)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating through rows: %v", err)
	}

	bar.Finish()
	log.Printf("Processed %d routes successfully", totalRoutes)
}

// Fetch airport coordinates from locations.db
func getAirportCoordinates(db *sql.DB, iataCode string) (float64, float64, error) {
	query := `SELECT lat, lon FROM airport WHERE iata = ?`
	var lat, lon float64
	err := db.QueryRow(query, iataCode).Scan(&lat, &lon)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to find coordinates for IATA code %s: %v", iataCode, err)
	}
	return lat, lon, nil
}
