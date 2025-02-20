package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const basePricePerMinute = 0.15

// k is the attenuation parameter for durations beyond 6 hours (360 minutes).
const k = 50.0

// Weighting factors for each modifier in log-space.
const (
	wAirline     = 9.0
	wPopulation  = 0.1
	wDate        = 0.1
	wFrequency   = 0.25
	wShortNotice = 0.1
	wCapacity    = 0.75
	wRouteClass  = 0.4
)

// effectiveDuration calculates the effective flight duration (in minutes) based on FD.
// For FD <= 120, it returns FD.
// For FD > 120, it returns 120 + k * ln(1 + (FD - 120)/k)
func effectiveDuration(fd int) int {
	if fd <= 80 {
		return fd
	}
	extra := float64(fd - 80)
	effectiveExtra := k * math.Log(1+extra/k)
	return 80 + int(math.Round(effectiveExtra))
}

// Route holds the fields we read from the routes table.
type Route struct {
	id                  int
	durationInMinutes   sql.NullInt64
	durationHourDotMins sql.NullString
	routeFrequency      int
	mostCommonAirline   sql.NullString
	originPopulation    int
	destPopulation      int
	aircraftCapacity    int
	routeClassification sql.NullString
}

// --- Lookup functions (same as before) ---

func lookupMultiplier(db *sql.DB, query string, param interface{}) float64 {
	var multiplier float64
	err := db.QueryRow(query, param).Scan(&multiplier)
	if err != nil {
		if err == sql.ErrNoRows {
			return 1.0
		}
		log.Printf("Lookup error (%s, param=%v): %v", query, param, err)
		return 1.0
	}
	return multiplier
}

func lookupPopulationModifier(db *sql.DB, population int) float64 {
	query := `
		SELECT multiplier FROM population_modifiers
		WHERE ? BETWEEN min_population AND max_population
		LIMIT 1;
	`
	var multiplier float64
	err := db.QueryRow(query, population).Scan(&multiplier)
	if err != nil {
		if err == sql.ErrNoRows {
			return 1.0
		}
		log.Printf("Population lookup error for population %d: %v", population, err)
		return 1.0
	}
	return multiplier
}

func lookupCapacityModifier(db *sql.DB, capacity int) float64 {
	query := `
		SELECT multiplier FROM aircraft_capacity_modifiers
		WHERE ? >= min_capacity AND (max_capacity IS NULL OR ? <= max_capacity)
		LIMIT 1;
	`
	var multiplier float64
	err := db.QueryRow(query, capacity, capacity).Scan(&multiplier)
	if err != nil {
		if err == sql.ErrNoRows {
			return 1.0
		}
		log.Printf("Capacity lookup error for capacity %d: %v", capacity, err)
		return 1.0
	}
	return multiplier
}

func lookupFlightFrequencyModifier(db *sql.DB, flights int) float64 {
	query := `
		SELECT multiplier FROM flight_frequency_modifiers
		WHERE ? BETWEEN min_flights AND max_flights
		LIMIT 1;
	`
	var multiplier float64
	err := db.QueryRow(query, flights).Scan(&multiplier)
	if err != nil {
		if err == sql.ErrNoRows {
			return 1.0
		}
		log.Printf("Flight frequency lookup error for %d flights: %v", flights, err)
		return 1.0
	}
	return multiplier
}

func lookupDateModifier(db *sql.DB) float64 {
	today := time.Now().Format("2006-01-02")
	query := `
		SELECT multiplier FROM date_modifiers
		WHERE start_date <= ? AND end_date >= ?
		LIMIT 1;
	`
	var multiplier float64
	err := db.QueryRow(query, today, today).Scan(&multiplier)
	if err != nil {
		if err == sql.ErrNoRows {
			return 1.0
		}
		log.Printf("Date modifier lookup error for date %s: %v", today, err)
		return 1.0
	}
	return multiplier
}

func lookupRouteClassificationModifier(db *sql.DB, classification string) float64 {
	query := `
		SELECT multiplier FROM route_classification_modifiers
		WHERE classification = ?
		LIMIT 1;
	`
	return lookupMultiplier(db, query, classification)
}

func main() {
	// Open the flight-prices database.
	fpDBPath := "../../../data/generated/flight-prices.db"
	fpDB, err := sql.Open("sqlite3", fpDBPath)
	if err != nil {
		log.Fatalf("Failed to open flight-prices database: %v", err)
	}
	defer fpDB.Close()
	_, err = fpDB.Exec("PRAGMA busy_timeout = 5000")
	if err != nil {
		log.Printf("Error setting busy_timeout: %v", err)
	}

	// Open the flight price modifiers database.
	modDBPath := "../../../data/generated/flight_price_modifiers.db"
	modDB, err := sql.Open("sqlite3", modDBPath)
	if err != nil {
		log.Fatalf("Failed to open flight_price_modifiers database: %v", err)
	}
	defer modDB.Close()
	_, err = modDB.Exec("PRAGMA busy_timeout = 5000")
	if err != nil {
		log.Printf("Error setting busy_timeout on modDB: %v", err)
	}

	// Read all routes into memory.
	rows, err := fpDB.Query(`
		SELECT id, duration_in_minutes, duration_hour_dot_mins, route_frequency, most_common_airline, 
		       origin_population, destination_population, most_common_aircraft_seating_capacity,
		       route_classification
		FROM routes;
	`)
	if err != nil {
		log.Fatalf("Failed to query routes: %v", err)
	}
	defer rows.Close()

	var routes []Route
	for rows.Next() {
		var r Route
		err := rows.Scan(&r.id, &r.durationInMinutes, &r.durationHourDotMins, &r.routeFrequency, &r.mostCommonAirline,
			&r.originPopulation, &r.destPopulation, &r.aircraftCapacity, &r.routeClassification)
		if err != nil {
			log.Printf("Error scanning route row: %v", err)
			continue
		}
		routes = append(routes, r)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating routes: %v", err)
	}

	// Begin a transaction for updates.
	tx, err := fpDB.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	updateStmt, err := tx.Prepare(`UPDATE routes SET calculated_price = ? WHERE id = ?;`)
	if err != nil {
		log.Fatalf("Failed to prepare update statement: %v", err)
	}
	defer updateStmt.Close()

	count := 0
	for _, r := range routes {
		// Calculate flight duration (FD) in minutes.
		var fd int
		if r.durationHourDotMins.Valid {
			parts := strings.Split(r.durationHourDotMins.String, ".")
			if len(parts) == 2 {
				hours, err1 := strconv.Atoi(parts[0])
				mins, err2 := strconv.Atoi(parts[1])
				if err1 != nil || err2 != nil {
					log.Printf("Skipping route id %d: unable to parse duration_hour_dot_mins (%s)", r.id, r.durationHourDotMins.String)
					continue
				}
				fd = hours*60 + mins*10
			} else {
				log.Printf("Skipping route id %d: unexpected duration_hour_dot_mins format (%s)", r.id, r.durationHourDotMins.String)
				continue
			}
		} else {
			log.Printf("Skipping route id %d: no duration_hour_dot_mins information", r.id)
			continue
		}

		// Apply non-linear adjustment to flight duration.
		effFD := effectiveDuration(fd) // Base fare = basePricePerMinute * effective duration
		baseFare := basePricePerMinute * float64(effFD) / 1.3

		// Airline Multiplier.
		airlineValue := "Unknown"
		if r.mostCommonAirline.Valid {
			airlineValue = r.mostCommonAirline.String
		}
		amQuery := `SELECT multiplier FROM airline_multipliers WHERE airline = ? LIMIT 1;`
		airlineMultiplier := lookupMultiplier(modDB, amQuery, airlineValue)

		// Population Modifier: average of origin and destination modifiers.
		popModOrigin := lookupPopulationModifier(modDB, r.originPopulation)
		popModDest := lookupPopulationModifier(modDB, r.destPopulation)
		populationModifier := (popModOrigin + popModDest) / 2.0

		// Date/Season/Holiday Modifier.
		dateModifier := lookupDateModifier(modDB)

		// Flight Frequency Modifier.
		frequencyModifier := lookupFlightFrequencyModifier(modDB, r.routeFrequency)

		// Short-Notice Modifier (default 1.0).
		shortNoticeModifier := 1.8

		// Aircraft Capacity Multiplier.
		capacityModifier := lookupCapacityModifier(modDB, r.aircraftCapacity)

		// Route Classification Multiplier.
		classificationValue := "Unknown"
		if r.routeClassification.Valid {
			classificationValue = r.routeClassification.String
		}
		routeClassModifier := lookupRouteClassificationModifier(modDB, classificationValue)

		// Calculate final price using a log-linear model.
		// (Make sure that none of the multipliers are <= 0; defaults are 1.0)
		logFinalPrice := math.Log(baseFare) +
			wAirline*math.Log(airlineMultiplier) +
			wPopulation*math.Log(populationModifier) +
			wDate*math.Log(dateModifier) +
			wFrequency*math.Log(frequencyModifier) +
			wShortNotice*math.Log(shortNoticeModifier) +
			wCapacity*math.Log(capacityModifier) +
			wRouteClass*math.Log(routeClassModifier)
		finalPrice := ((math.Exp(logFinalPrice) * 10) + 40) //* airlineMultiplier

		var maxPricePerMinute float64

		// if effFD < 240 { // under 4 hours
		// 	maxPricePerMinute = 4.0
		// } else if effFD < 480 { // between 4 and 8 hours
		// 	maxPricePerMinute = 3.0
		// } else { // over 8 hours
		// 	maxPricePerMinute = 2.5
		// }
		maxPricePerMinute = 2.1
		// Calculate the computed price per minute.
		pricePerMinute := finalPrice / float64(fd)
		// If it exceeds the cap, reduce the final price.
		if pricePerMinute > maxPricePerMinute {
			finalPrice = maxPricePerMinute * float64(fd)
			randomMultiplier := rand.Float64()*(1.1-0.9) + 0.9
			finalPrice = finalPrice * randomMultiplier

		}

		if fd > 240 {
			minPricePerMinute := 1.4
			pricePerMinute = finalPrice / float64(fd)
			if pricePerMinute < minPricePerMinute {
				finalPrice = minPricePerMinute * float64(fd)
				randomMultiplier := rand.Float64()*(1.1-0.9) + 0.9
				finalPrice = finalPrice * randomMultiplier

			}
		}

		//		finalPrice = finalPrice * airlineMultiplier
		// Update the route's calculated_price.
		_, err = updateStmt.Exec(finalPrice, r.id)
		if err != nil {
			log.Printf("Failed to update calculated_price for route id %d: %v", r.id, err)
			continue
		}
		count++
		log.Printf("Updated route id %d: fd=%d, baseFare=%.2f, AM=%.2f, PM=%.2f, DM=%.2f, FFM=%.2f, ACM=%.2f, RCM=%.2f => price=%.2f",
			r.id, fd, baseFare, airlineMultiplier, populationModifier, dateModifier, frequencyModifier, capacityModifier, routeClassModifier, finalPrice)
	}

	// Commit the transaction.
	err = tx.Commit()
	if err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Printf("Successfully updated calculated_price for %d routes.\n", count)
}
