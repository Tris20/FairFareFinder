package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sajari/regression"
)

// Encoding maps and helper functions.
var (
	originCityMap      = make(map[string]float64)
	destinationCityMap = make(map[string]float64)
	airlineMap         = make(map[string]float64)
	routeClassMap      = make(map[string]float64)
	aircraftMap        = make(map[string]float64)

	nextOriginCityCode float64 = 1
	nextDestCityCode   float64 = 1
	nextAirlineCode    float64 = 1
	nextRouteClassCode float64 = 1
	nextAircraftCode   float64 = 1
)

// encodeString returns a numeric encoding for a given category.
func encodeString(val string, m map[string]float64, next *float64) float64 {
	if code, ok := m[val]; ok {
		return code
	}
	m[val] = *next
	*next++
	return m[val]
}

// parseDuration converts a "H.MM" string into total minutes.
// It calculates: totalMinutes = hours*60 + minutes*10.
// If no dot is present, the value is assumed to represent hours.
func parseDuration(duration string) (int, error) {
	if !strings.Contains(duration, ".") {
		hours, err := strconv.Atoi(duration)
		if err != nil {
			return 0, fmt.Errorf("invalid hours value in duration: %s", duration)
		}
		return hours * 60, nil
	}
	parts := strings.Split(duration, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("unexpected duration format: %s", duration)
	}
	hours, err1 := strconv.Atoi(parts[0])
	mins, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, fmt.Errorf("invalid numeric values in duration: %s", duration)
	}
	return hours*60 + mins*10, nil
}

func main() {
	// Create the generated flight price db and routes table
	CreateResultsDB()
	// Populate the Routes table with all the flight data except prices
	PopulateRoutesTable()
	// Train on dataset and then generate predicted flight prices into predcitions table
	GeneratePredictions()
}

func GeneratePredictions() {
	// Seed the random number generator.
	rand.Seed(time.Now().UnixNano())

	// ============================================
	// PART 1: TRAINING PHASE – Build Regression Model
	// ============================================
	// Open the training CSV file.
	f, err := os.Open("training_data.csv")
	if err != nil {
		log.Fatalf("Error opening CSV file: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV records: %v", err)
	}

	// Create and configure the regression model.
	var regModel regression.Regression
	regModel.SetObserved("actual_price")
	regModel.SetVar(0, "bias")
	regModel.SetVar(1, "origin_population")
	regModel.SetVar(2, "destination_population")
	regModel.SetVar(3, "route_frequency")
	regModel.SetVar(4, "origin_city")
	regModel.SetVar(5, "destination_city")
	regModel.SetVar(6, "airline")
	regModel.SetVar(7, "route_class")
	regModel.SetVar(8, "aircraft")
	regModel.SetVar(9, "seating_capacity")
	// Changed variable name from "duration_minutes" to "duration_hour_dot_minutes"
	regModel.SetVar(10, "duration_hour_dot_minutes")

	startRow := 0
	if strings.Contains(strings.ToLower(records[0][0]), "origin_city_name") {
		startRow = 1
	}

	// Train the regression model using CSV records.
	for i := startRow; i < len(records); i++ {
		rec := records[i]
		if len(rec) < 15 {
			continue
		}
		originPop, _ := strconv.Atoi(rec[3])
		destPop, _ := strconv.Atoi(rec[7])
		routeFreq, _ := strconv.Atoi(rec[8])
		seatCapacity, _ := strconv.Atoi(rec[12])
		durationStr := rec[13]
		actualPrice, err := strconv.ParseFloat(rec[14], 64)
		if err != nil {
			log.Printf("Skipping record %d due to invalid price: %v", i, err)
			continue
		}
		// Convert the duration string (H.MM format) to minutes.
		durationMinutes, err := parseDuration(durationStr)
		if err != nil {
			log.Printf("Skipping record %d: %v", i, err)
			continue
		}

		originCityEnc := encodeString(rec[0], originCityMap, &nextOriginCityCode)
		destCityEnc := encodeString(rec[4], destinationCityMap, &nextDestCityCode)
		airlineEnc := encodeString(rec[10], airlineMap, &nextAirlineCode)
		routeClassEnc := encodeString(rec[9], routeClassMap, &nextRouteClassCode)
		aircraftEnc := encodeString(rec[11], aircraftMap, &nextAircraftCode)

		features := []float64{
			1.0,
			float64(originPop),
			float64(destPop),
			float64(routeFreq),
			originCityEnc,
			destCityEnc,
			airlineEnc,
			routeClassEnc,
			aircraftEnc,
			float64(seatCapacity),
			float64(durationMinutes),
		}
		regModel.Train(regression.DataPoint(actualPrice, features))
	}
	regModel.Run()
	fmt.Printf("Regression Formula:\n%v\n\n", regModel.Formula)

	// ====================================================
	// PART 2: PREDICTION REFINEMENT – Using CSV Records With Actual Price
	// (Generate a table with full data where actual_price exists)
	// ====================================================
	// Open the predictions database (flight-prices.db in the generated folder).
	db, err := sql.Open("sqlite3", "../../../../../data/generated/flight-prices.db")
	if err != nil {
		log.Fatalf("Error opening flight-prices.db: %v", err)
	}
	defer db.Close()

	// Create the prediction_refinement table.
	createRefinementTableSQL := `
CREATE TABLE IF NOT EXISTS prediction_refinement (
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
	duration_hour_dot_mins TEXT,
	actual_price REAL,
	predicted_price REAL,
	price_difference REAL,
	error_multiple REAL,
	error_direction TEXT
);
`
	_, err = db.Exec(createRefinementTableSQL)
	if err != nil {
		log.Fatalf("Error creating prediction_refinement table: %v", err)
	}

	// Prepare an INSERT statement for prediction_refinement.
	insertRefinementStmt, err := db.Prepare(`
INSERT INTO prediction_refinement (
	origin_city_name, origin_country, origin_iata, origin_population,
	destination_city_name, destination_country, destination_iata, destination_population,
	route_frequency, route_classification, most_common_airline, most_common_aircraft,
	most_common_aircraft_seating_capacity, duration_hour_dot_mins,
	actual_price, predicted_price, price_difference, error_multiple, error_direction
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
`)
	if err != nil {
		log.Fatalf("Error preparing insert statement for prediction_refinement: %v", err)
	}
	defer insertRefinementStmt.Close()

	// Iterate over CSV records (with actual prices) to fill prediction_refinement.
	for i := startRow; i < len(records); i++ {
		rec := records[i]
		if len(rec) < 15 {
			continue
		}

		originPop, _ := strconv.Atoi(rec[3])
		destPop, _ := strconv.Atoi(rec[7])
		routeFreq, _ := strconv.Atoi(rec[8])
		seatCapacity, _ := strconv.Atoi(rec[12])
		durationStr := rec[13]
		actualPrice, err := strconv.ParseFloat(rec[14], 64)
		if err != nil {
			log.Printf("Skipping record %d due to invalid price: %v", i, err)
			continue
		}
		durationMinutes, err := parseDuration(durationStr)
		if err != nil {
			log.Printf("Skipping record %d: %v", i, err)
			continue
		}

		originCityEnc := encodeString(rec[0], originCityMap, &nextOriginCityCode)
		destCityEnc := encodeString(rec[4], destinationCityMap, &nextDestCityCode)
		airlineEnc := encodeString(rec[10], airlineMap, &nextAirlineCode)
		routeClassEnc := encodeString(rec[9], routeClassMap, &nextRouteClassCode)
		aircraftEnc := encodeString(rec[11], aircraftMap, &nextAircraftCode)

		features := []float64{
			1.0,
			float64(originPop),
			float64(destPop),
			float64(routeFreq),
			originCityEnc,
			destCityEnc,
			airlineEnc,
			routeClassEnc,
			aircraftEnc,
			float64(seatCapacity),
			float64(durationMinutes),
		}

		predictedPrice, _ := regModel.Predict(features)
		finalPrice := predictedPrice

		// Apply boundary rules.
		fd := durationMinutes
		pricePerMinute := finalPrice / float64(fd)
		randomMultiplier := rand.Float64()*(1.1-0.9) + 0.9
		if fd > 240 {
			randomMultiplier = rand.Float64()*(1.05-0.95) + 0.95
			if pricePerMinute < 1.4 {
				finalPrice = 1.4 * float64(fd) * randomMultiplier
			}
			if pricePerMinute > 2.3 {
				finalPrice = 2.3 * float64(fd) * randomMultiplier
			}
		} else if fd <= 60 {
			if pricePerMinute < 0.9 {
				finalPrice = 60.0 * randomMultiplier
			}
			if finalPrice > 210 {
				finalPrice = 0
			}
		} else if fd > 60 && fd <= 120 {
			if pricePerMinute < 0.9 {
				finalPrice = 0.9 * float64(fd) * randomMultiplier
			}
			if finalPrice > 240 {
				finalPrice = 0
			}
		} else if fd > 120 && fd <= 240 {
			if pricePerMinute < 0.9 {
				finalPrice = 0.9 * float64(fd) * randomMultiplier
			}
			if finalPrice > 260 {
				finalPrice = 0
			}
		}

		priceDifference := finalPrice - actualPrice
		var errorMultiple float64
		if actualPrice > 0 && finalPrice > 0 {
			if finalPrice >= actualPrice {
				errorMultiple = finalPrice / actualPrice
			} else {
				errorMultiple = actualPrice / finalPrice
			}
		}
		var errorDirection string
		if finalPrice > actualPrice {
			errorDirection = "too high"
		} else if finalPrice < actualPrice {
			errorDirection = "too low"
		} else {
			errorDirection = "equal"
		}

		_, err = insertRefinementStmt.Exec(
			rec[0], // origin_city_name
			rec[1], // origin_country
			rec[2], // origin_iata
			originPop,
			rec[4], // destination_city_name
			rec[5], // destination_country
			rec[6], // destination_iata
			destPop,
			routeFreq,
			rec[9],  // route_classification
			rec[10], // most_common_airline
			rec[11], // most_common_aircraft
			seatCapacity,
			rec[13], // duration_hour_dot_mins (text value from CSV)
			actualPrice,
			finalPrice,
			priceDifference,
			errorMultiple,
			errorDirection,
		)
		if err != nil {
			log.Printf("Insert error in prediction_refinement on record %d: %v", i, err)
		}
	}

	// ============================================================
	// PART 3: PREDICTION – Generate Predictions for Every Route
	// (Using the "routes" table, even when no actual price exists)
	// ============================================================
	// Create the prediction table.
	createPredictionTableSQL := `
CREATE TABLE IF NOT EXISTS prediction (
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
	duration_hour_dot_mins TEXT,
	predicted_price REAL
);
`
	_, err = db.Exec(createPredictionTableSQL)
	if err != nil {
		log.Fatalf("Error creating prediction table: %v", err)
	}

	// Begin a transaction for prediction inserts.
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Error beginning transaction: %v", err)
	}

	insertPredictionStmt, err := tx.Prepare(`
INSERT INTO prediction (
	origin_city_name, origin_country, origin_iata, origin_population,
	destination_city_name, destination_country, destination_iata, destination_population,
	route_frequency, route_classification, most_common_airline, most_common_aircraft,
	most_common_aircraft_seating_capacity, duration_hour_dot_mins,
	predicted_price
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
`)
	if err != nil {
		log.Fatalf("Error preparing insert statement for prediction: %v", err)
	}
	defer insertPredictionStmt.Close()

	// Query every route from the routes table.
	// We now select the text column "duration_hour_dot_mins".
	routesRows, err := db.Query(`SELECT origin_city_name, origin_country, origin_iata, origin_population,
		destination_city_name, destination_country, destination_iata, destination_population,
		route_frequency, route_classification, most_common_airline, most_common_aircraft,
		most_common_aircraft_seating_capacity, duration_hour_dot_mins
		FROM routes;`)
	if err != nil {
		log.Fatalf("Error querying routes table: %v", err)
	}
	defer routesRows.Close()

	for routesRows.Next() {
		var originCity, originCountry, originIATA string
		var originPopulation int
		var destCity, destCountry, destIATA string
		var destPopulation int
		var routeFreq int
		var routeClass string
		var commonAirline string
		var commonAircraft string
		var seatingCapacity int
		var durationHdotMins sql.NullString

		err = routesRows.Scan(&originCity, &originCountry, &originIATA, &originPopulation,
			&destCity, &destCountry, &destIATA, &destPopulation,
			&routeFreq, &routeClass, &commonAirline, &commonAircraft,
			&seatingCapacity, &durationHdotMins)
		if err != nil {
			log.Printf("Error scanning route row: %v", err)
			continue
		}

		// Parse the duration from the text column.
		var durationMinutes int
		if durationHdotMins.Valid && durationHdotMins.String != "" {
			durationMinutes, err = parseDuration(durationHdotMins.String)
			if err != nil {
				log.Printf("Error parsing duration for route %s -> %s: %v", originIATA, destIATA, err)
				// Use a fallback default (e.g., 120 minutes)
				durationMinutes = 120
			}
		} else {
			// Fallback default if value is NULL or empty.
			durationMinutes = 120
		}

		// Build the feature vector.
		originCityEnc := encodeString(originCity, originCityMap, &nextOriginCityCode)
		destCityEnc := encodeString(destCity, destinationCityMap, &nextDestCityCode)
		airlineEnc := encodeString(commonAirline, airlineMap, &nextAirlineCode)
		routeClassEnc := encodeString(routeClass, routeClassMap, &nextRouteClassCode)
		aircraftEnc := encodeString(commonAircraft, aircraftMap, &nextAircraftCode)

		features := []float64{
			1.0,
			float64(originPopulation),
			float64(destPopulation),
			float64(routeFreq),
			originCityEnc,
			destCityEnc,
			airlineEnc,
			routeClassEnc,
			aircraftEnc,
			float64(seatingCapacity),
			float64(durationMinutes),
		}

		predictedPrice, _ := regModel.Predict(features)
		finalPrice := predictedPrice

		// Apply the same boundary rules.
		fd := durationMinutes
		pricePerMinute := finalPrice / float64(fd)
		randomMultiplier := rand.Float64()*(1.1-0.9) + 0.9
		if fd > 240 {
			randomMultiplier = rand.Float64()*(1.05-0.95) + 0.95
			if pricePerMinute < 1.4 {
				finalPrice = 1.4 * float64(fd) * randomMultiplier
			}
			if pricePerMinute > 2.3 {
				finalPrice = 2.3 * float64(fd) * randomMultiplier
			}
		} else if fd <= 60 {
			if pricePerMinute < 0.9 {
				finalPrice = 60.0 * randomMultiplier
			}
			if finalPrice > 210 {
				finalPrice = 0
			}
		} else if fd > 60 && fd <= 120 {
			if pricePerMinute < 0.9 {
				finalPrice = 0.9 * float64(fd) * randomMultiplier
			}
			if finalPrice > 240 {
				finalPrice = 0
			}
		} else if fd > 120 && fd <= 240 {
			if pricePerMinute < 0.9 {
				finalPrice = 0.9 * float64(fd) * randomMultiplier
			}
			if finalPrice > 260 {
				finalPrice = 0
			}
		}

		// Use the original text value (or empty string) for duration_hour_dot_mins.
		durStr := ""
		if durationHdotMins.Valid {
			durStr = durationHdotMins.String
		}

		_, err = insertPredictionStmt.Exec(
			originCity,
			originCountry,
			originIATA,
			originPopulation,
			destCity,
			destCountry,
			destIATA,
			destPopulation,
			routeFreq,
			routeClass,
			commonAirline,
			commonAircraft,
			seatingCapacity,
			durStr,
			finalPrice,
		)
		if err != nil {
			log.Printf("Insert error in prediction for route %s -> %s: %v", originIATA, destIATA, err)
		}
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		log.Fatalf("Error committing transaction: %v", err)
	}

	fmt.Println("prediction_refinement and prediction tables updated in flight-prices.db")
}
