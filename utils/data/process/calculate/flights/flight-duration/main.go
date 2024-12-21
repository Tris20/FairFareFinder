package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"

	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
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
	// Paths to the databases
	mainDBPath := "../../../../../../data/compiled/new_main.db"
	locationsDBPath := "../../../../../../data/raw/locations/locations.db"

	log.Printf("Connecting to main database at: %s", mainDBPath)
	mainDB, err := sql.Open("sqlite3", mainDBPath)
	if err != nil {
		log.Fatalf("Failed to open main.db: %v", err)
	}
	defer mainDB.Close()

	log.Printf("Connecting to locations database at: %s", locationsDBPath)
	locationsDB, err := sql.Open("sqlite3", locationsDBPath)
	if err != nil {
		log.Fatalf("Failed to open locations.db: %v", err)
	}
	defer locationsDB.Close()

	// Query flight routes and process them
	log.Printf("Querying flight routes from flight table")
	flightQuery := `SELECT id, origin_iata, destination_iata FROM flight`
	rows, err := mainDB.Query(flightQuery)
	if err != nil {
		log.Fatalf("Failed to query flight table: %v", err)
	}
	defer rows.Close()

	// Start transaction for updates
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

	// Initialize progress bar
	var totalFlights int
	bar := progressbar.NewOptions(-1, progressbar.OptionSetDescription("Calculating flight durations"))

	// Iterate through flight routes
	for rows.Next() {
		totalFlights++
		var flightID int
		var originIATA, destinationIATA string
		if err := rows.Scan(&flightID, &originIATA, &destinationIATA); err != nil {
			log.Fatalf("Failed to scan flight row: %v", err)
		}

		log.Printf("Processing flight ID %d: %s -> %s", flightID, originIATA, destinationIATA)

		// Get lat/lon for origin airport
		originLat, originLon, err := getAirportCoordinates(locationsDB, originIATA)
		if err != nil {
			log.Printf("Skipping flight ID %d due to missing origin coordinates: %v", flightID, err)
			bar.Add(1)
			continue
		}

		// Get lat/lon for destination airport
		destLat, destLon, err := getAirportCoordinates(locationsDB, destinationIATA)
		if err != nil {
			log.Printf("Skipping flight ID %d due to missing destination coordinates: %v", flightID, err)
			bar.Add(1)
			continue
		}

		// Calculate distance and flight time
		distance := haversine(originLat, originLon, destLat, destLon)
		log.Printf("Calculated distance for flight ID %d: %.2f km", flightID, distance)
		flightTime := calculateAdjustedFlightTime(distance)

		log.Printf("Flight ID %d: Calculated flightTime = %.2f hours", flightID, flightTime)
		// Calculate formatted durations
		durationHourDotMins, durationMinutes, durationHoursRounded := formatDuration(flightTime)
		hoursPart := int(durationHourDotMins) // Hours without minutes

		// Log calculated values
		log.Printf("Flight ID %d: DurationHourDotMins = %.2f, DurationInMinutes = %d, DurationInHours = %d, DurationHoursRounded = %.2f",
			flightID, durationHourDotMins, durationMinutes, hoursPart, durationHoursRounded)

		// Update flight duration within the transaction
		updateQuery := `
			UPDATE flight 
			SET duration_hour_dot_mins = ?, 
			    duration_in_minutes = ?, 
			    duration_in_hours = ?, 
			    duration_in_hours_rounded = ? 
			WHERE id = ?`
		_, err = tx.Exec(updateQuery, durationHourDotMins, durationMinutes, float64(hoursPart), durationHoursRounded, flightID)
		if err != nil {
			log.Printf("Failed to update flight duration for ID %d: %v", flightID, err)
		} else {
			log.Printf("Successfully updated flight ID %d", flightID)
		}

		bar.Add(1)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating through flight rows: %v", err)
	}

	bar.Finish()
	log.Printf("Processed %d flight routes successfully", totalFlights)
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
