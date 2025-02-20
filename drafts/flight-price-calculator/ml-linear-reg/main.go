package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sajari/regression"
)

// (Encoding maps and helper functions from previous code...)
var (
	originCityMap              = make(map[string]float64)
	destinationCityMap         = make(map[string]float64)
	airlineMap                 = make(map[string]float64)
	routeClassMap              = make(map[string]float64)
	aircraftMap                = make(map[string]float64)
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

// parseDuration converts a "H.MM" string (e.g. "7.45") into total minutes.
// If no dot is present, we assume the value represents hours.
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
	minutes, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, fmt.Errorf("invalid numeric values in duration: %s", duration)
	}
	return hours*60 + minutes, nil
}

func main() {
	// --- Model Training Section (same as before) ---

	// Open the training CSV file.
	f, err := os.Open("flight_prices.csv")
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
	var r regression.Regression
	r.SetObserved("actual_price")
	r.SetVar(0, "bias")
	r.SetVar(1, "origin_population")
	r.SetVar(2, "destination_population")
	r.SetVar(3, "route_frequency")
	r.SetVar(4, "origin_city")
	r.SetVar(5, "destination_city")
	r.SetVar(6, "airline")
	r.SetVar(7, "route_class")
	r.SetVar(8, "aircraft")
	r.SetVar(9, "seating_capacity")
	r.SetVar(10, "duration_minutes")

	startRow := 0
	if strings.Contains(strings.ToLower(records[0][0]), "origin_city_name") {
		startRow = 1
	}

	// Train the model.
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
			float64(aircraftEnc),
			float64(seatCapacity),
			float64(durationMinutes),
		}
		r.Train(regression.DataPoint(actualPrice, features))
	}
	r.Run()
	fmt.Printf("Regression Formula:\n%v\n\n", r.Formula)

	// --- Prediction and Output Section ---

	// Open (or create) the predictions SQLite DB.
	predDB, err := os.OpenFile("predictions.db", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("Failed to open predictions.db: %v", err)
	}
	predDB.Close() // We'll use the sqlite3 driver to create the DB below.

	// Open a connection using the sqlite3 driver.
	db, err := sql.Open("sqlite3", "predictions.db")
	if err != nil {
		log.Fatalf("Error opening predictions.db: %v", err)
	}
	defer db.Close()

	// Create the predictions table.
	createTableSQL := `
CREATE TABLE IF NOT EXISTS predictions (
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
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Error creating predictions table: %v", err)
	}

	// Prepare an insert statement.
	insertStmt, err := db.Prepare(`
INSERT INTO predictions (
	origin_city_name, origin_country, origin_iata, origin_population,
	destination_city_name, destination_country, destination_iata, destination_population,
	route_frequency, route_classification, most_common_airline, most_common_aircraft,
	most_common_aircraft_seating_capacity, duration_hour_dot_mins,
	actual_price, predicted_price, price_difference, error_multiple, error_direction
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
`)
	if err != nil {
		log.Fatalf("Error preparing insert statement: %v", err)
	}
	defer insertStmt.Close()

	// Iterate over the CSV records again, predict and insert results.
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
		// Encode the same categorical fields.
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
			float64(aircraftEnc),
			float64(seatCapacity),
			float64(durationMinutes),
		}

		// Get the predicted price.
		predictedPrice, _ := r.Predict(features)
		priceDifference := predictedPrice - actualPrice

		var errorMultiple float64
		if actualPrice > 0 && predictedPrice > 0 {
			if predictedPrice >= actualPrice {
				errorMultiple = predictedPrice / actualPrice
			} else {
				errorMultiple = actualPrice / predictedPrice
			}
		}

		var errorDirection string
		if predictedPrice > actualPrice {
			errorDirection = "too high"
		} else if predictedPrice < actualPrice {
			errorDirection = "too low"
		} else {
			errorDirection = "equal"
		}

		// Insert all fields into the predictions table.
		_, err = insertStmt.Exec(
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
			rec[13], // duration_hour_dot_mins
			actualPrice,
			predictedPrice,
			priceDifference,
			errorMultiple,
			errorDirection,
		)
		if err != nil {
			log.Printf("Insert error on record %d: %v", i, err)
		}
	}

	fmt.Println("Predictions inserted into predictions.db table 'predictions'")
}
