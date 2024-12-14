package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"

	_ "github.com/mattn/go-sqlite3"
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

func main() {
	// Create or open the SQLite database
	dbPath := "./routes.db"
	os.Remove(dbPath) // Remove existing database for a fresh start
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the routes table
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS routes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		origin_city TEXT NOT NULL,
		origin_lat REAL NOT NULL,
		origin_lon REAL NOT NULL,
		destination_city TEXT NOT NULL,
		destination_lat REAL NOT NULL,
		destination_lon REAL NOT NULL,
		distance_km REAL NOT NULL,
		estimated_flight_time_hr REAL NOT NULL
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// Example data
	routes := []struct {
		originCity           string
		originLat, originLon float64
		destinationCity      string
		destLat, destLon     float64
	}{
		{"New York", 40.7128, -74.0060, "London", 51.5074, -0.1278},
		{"San Francisco", 37.7749, -122.4194, "Tokyo", 35.6895, 139.6917},
		{"Sydney", -33.8688, 151.2093, "Auckland", -36.8485, 174.7633},
		{"Los Angeles", 34.0522, -118.2437, "Paris", 48.8566, 2.3522},
		{"Dubai", 25.276987, 55.296249, "Mumbai", 19.0760, 72.8777},
		{"Cape Town", -33.9249, 18.4241, "Johannesburg", -26.2041, 28.0473},
		{"Toronto", 43.65107, -79.347015, "Vancouver", 49.2827, -123.1207},
		{"Buenos Aires", -34.6037, -58.3816, "Santiago", -33.4489, -70.6693},
		{"Cairo", 30.0444, 31.2357, "Rome", 41.9028, 12.4964},
		{"Singapore", 1.3521, 103.8198, "Bangkok", 13.7563, 100.5018},
		{"Beijing", 39.9042, 116.4074, "Seoul", 37.5665, 126.9780},
		{"New Delhi", 28.6139, 77.2090, "Kathmandu", 27.7172, 85.3240},
		{"Berlin", 52.5200, 13.4050, "Amsterdam", 52.3676, 4.9041},
		{"Mexico City", 19.4326, -99.1332, "Miami", 25.7617, -80.1918},
	}

	// Insert data into the routes table
	insertQuery := `
	INSERT INTO routes (origin_city, origin_lat, origin_lon, destination_city, destination_lat, destination_lon, distance_km, estimated_flight_time_hr)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?);`
	for _, route := range routes {
		distance := haversine(route.originLat, route.originLon, route.destLat, route.destLon)
		flightTime := calculateAdjustedFlightTime(distance)
		_, err := db.Exec(insertQuery, route.originCity, route.originLat, route.originLon, route.destinationCity, route.destLat, route.destLon, distance, flightTime)
		if err != nil {
			log.Fatalf("Failed to insert route: %v", err)
		}
	}

	// Display the routes table
	rows, err := db.Query("SELECT origin_city, destination_city, distance_km, estimated_flight_time_hr FROM routes")
	if err != nil {
		log.Fatalf("Failed to query routes: %v", err)
	}
	defer rows.Close()

	fmt.Println("Routes Table:")
	for rows.Next() {
		var originCity, destinationCity string
		var distance, flightTime float64
		err = rows.Scan(&originCity, &destinationCity, &distance, &flightTime)
		if err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		fmt.Printf("%s to %s: %.2f km, %.2f hours\n", originCity, destinationCity, distance, flightTime)
	}
}
