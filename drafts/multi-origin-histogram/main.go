package main

import (
	"database/sql"
	"fmt"
	"github.com/Tris20/FairFareFinder/src/backend"
	"github.com/Tris20/FairFareFinder/src/backend/model"
	"log"
	"math"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// FilterInput represents the parsed filter inputs.
type FilterInput struct {
	Cities                []string
	LogicalOperators      []string
	MaxFlightPrices       []float64
	MaxAccommodationPrice float64
	OrderClause           string
	// LogicalExpression can be built from the arrays if needed.
}

// FlightResult is used internally to hold query rows.
type FlightResult struct {
	DestinationCity    string
	DestinationCountry string
	FlightPrice        float64
	ActiveOrigin       string
}

// adjustExpressionForActive recursively traverses the Expression tree and,
// if a CityCondition matches the active origin (by Name and, if available, Country),
// replaces its PriceLimit with globalThreshold. If the parsed condition’s Country is empty,
// we assume it should match.
func adjustExpressionForActive(expr backend.Expression, active backend.CityInput, globalThreshold float64) backend.Expression {
	switch e := expr.(type) {
	case *backend.CityCondition:
		if e.City.Name == active.Name && (e.City.Country == active.Country || e.City.Country == "") {
			return &backend.CityCondition{
				City: backend.CityInput{
					Name:       e.City.Name,
					Country:    active.Country, // set explicitly
					PriceLimit: globalThreshold,
				},
			}
		}
		return e
	case *backend.LogicalExpression:
		return &backend.LogicalExpression{
			Operator: e.Operator,
			Left:     adjustExpressionForActive(e.Left, active, globalThreshold),
			Right:    adjustExpressionForActive(e.Right, active, globalThreshold),
		}
	default:
		log.Fatalf("adjustExpressionForActive: unknown expression type")
		return nil
	}
}

// buildQueryForActiveOrigin constructs the final SQL query for a given active origin.
// It uses the backend package to parse the logical expression from the input arrays,
// adjusts the expression (so that for the active origin the threshold is globalThreshold),
// builds a destination set subquery from that expression (using INTERSECT/UNION logic),
// wraps it in a WITH clause, and finally joins it with the flight table (restricted to the active origin)
// plus weather, location, and accommodation.
// An extra column, ActiveOrigin, is added so that the result rows are marked with the active origin.
func buildQueryForActiveOrigin(active backend.CityInput, cities []string, logicalOperators []string, maxPrices []float64, globalThreshold, accomLimit float64) (string, []interface{}, error) {
	// 1. Parse the logical expression from the inputs.
	expr, err := backend.ParseLogicalExpression(cities, logicalOperators, maxPrices)
	if err != nil {
		return "", nil, fmt.Errorf("parsing expression: %w", err)
	}

	// 2. Adjust the expression so that any CityCondition matching the active origin uses the global threshold.
	adjustedExpr := adjustExpressionForActive(expr, active, globalThreshold)

	// 3. Build the destination set subquery from the adjusted expression.
	subquery, subArgs := backend.BuildFlightOriginsSubquery(adjustedExpr)

	// 4. Wrap the subquery in a WITH clause.
	withClause := fmt.Sprintf("WITH DestinationSet AS (\n%s\n)", subquery)

	// 5. Build the main query.
	// We add an extra column "? AS ActiveOrigin" so that each row is tagged with the active origin.
	mainQuery := `
SELECT DISTINCT 
  ds.destination_city_name,
  ds.destination_country,
  f.price_next_week AS FlightPrice,
  ? AS ActiveOrigin
FROM DestinationSet ds
JOIN flight f 
  ON ds.destination_city_name = f.destination_city_name
  AND ds.destination_country = f.destination_country
JOIN weather w 
  ON ds.destination_city_name = w.city 
  AND ds.destination_country = w.country
JOIN location l 
  ON ds.destination_city_name = l.city 
  AND ds.destination_country = l.country
LEFT JOIN accommodation a 
  ON ds.destination_city_name = a.city
  AND ds.destination_country = a.country
WHERE f.origin_city_name = ? 
  AND f.origin_country = ?
  AND f.price_next_week < ?
  AND a.booking_pppn IS NOT NULL
  AND a.booking_pppn <= ?
`
	// 6. Combine the WITH clause and main query.
	finalQuery := withClause + "\n" + mainQuery

	// 7. Build the full list of arguments.
	// The placeholders appear in the following order:
	// - All placeholders from the subquery (subArgs)
	// - One for the SELECT ActiveOrigin column (active.Name)
	// - Then for the WHERE clause: active.Name, active.Country, globalThreshold, accomLimit.
	fullArgs := append(subArgs, active.Name, active.Name, active.Country, globalThreshold, accomLimit)

	return finalQuery, fullArgs, nil
}

// ExecuteFlightHistogramQuery builds and executes the query using the given FilterInput
// and returns an array of model.Flight.
func ExecuteFlightHistogramQuery(input *FilterInput) ([]model.Flight, error) {
	// Open the database.
	db, err := sql.Open("sqlite3", "../../data/compiled/histogram_test.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}
	defer db.Close()

	// Load the city-country pairs (if not already loaded) from the DB.
	backend.LoadCityCountryPairs(db)
	cityCountryPairs := backend.GetCityCountryPairs()

	// Derive active origins using input.Cities and input.MaxFlightPrices.
	var activeOrigins []backend.CityInput
	for i, city := range input.Cities {
		country := ""
		for _, cc := range cityCountryPairs {
			if cc.City == city {
				country = cc.Country
				break
			}
		}
		if country == "" {
			log.Printf("Warning: no country found for city %s", city)
		}
		activeOrigins = append(activeOrigins, backend.CityInput{
			Name:       city,
			Country:    country,
			PriceLimit: input.MaxFlightPrices[i],
		})
	}

	// Global flight price threshold.
	globalFlightPriceThreshold := 2500.00

	// Prepare to collect flights.
	var flights []model.Flight

	// Loop over each active origin.
	for _, active := range activeOrigins {
		query, args, err := buildQueryForActiveOrigin(active, input.Cities, input.LogicalOperators, input.MaxFlightPrices, globalFlightPriceThreshold, input.MaxAccommodationPrice)
		if err != nil {
			return nil, fmt.Errorf("error building query for active origin %s: %w", active.Name, err)
		}
		// Append the OrderClause from the FilterInput if provided.
		if strings.TrimSpace(input.OrderClause) != "" {
			query = query + "\n" + input.OrderClause
		}
		// log.Printf("Query for active origin %s:\n%s\nArgs: %v", active.Name, query, args)
		rows, err := db.Query(query, args...)
		if err != nil {
			return nil, fmt.Errorf("DB query failed for active origin %s: %w", active.Name, err)
		}
		for rows.Next() {
			var destCity, destCountry, activeOrigin string
			var price float64
			if err := rows.Scan(&destCity, &destCountry, &price, &activeOrigin); err != nil {
				rows.Close()
				return nil, fmt.Errorf("row scan failed for active origin %s: %w", active.Name, err)
			}
			// Round price to 2 decimal places:
			price = math.Round(price*100) / 100
			// Create a model.Flight record.
			flight := model.Flight{
				DestinationCityName: destCity,
				PriceCity1:          sql.NullFloat64{Float64: price, Valid: true},
			}
			// For demonstration, we store the active origin in UrlCity1.
			flight.UrlCity1 = activeOrigin
			flights = append(flights, flight)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("rows error for active origin %s: %w", active.Name, err)
		}
		rows.Close()
	}

	return flights, nil
}

func main() {
	// Example FilterInput.
	// Test 1 AND OR
	input := &FilterInput{
		Cities:                []string{"Berlin", "Glasgow", "Frankfurt"},
		LogicalOperators:      []string{"AND", "OR"},
		MaxFlightPrices:       []float64{122.00, 113.00, 96.00},
		MaxAccommodationPrice: 56.00,
		OrderClause:           "ORDER BY f.price_next_week ASC",
	}
	// Test 2 AND AND
	// input := &FilterInput{
	// 	Cities:                []string{"Berlin", "Glasgow", "Frankfurt"},
	// 	LogicalOperators:      []string{"AND", "AND"},
	// 	MaxFlightPrices:       []float64{115.00, 107.00, 177.00},
	// 	MaxAccommodationPrice: 150.00,
	// 	OrderClause:           "ORDER BY f.price_next_week ASC",
	// }
	// Test 3 Single City
	// input := &FilterInput{
	// 	Cities:                []string{"Glasgow"},
	// 	MaxFlightPrices:       []float64{115.00},
	// 	MaxAccommodationPrice: 150.00,
	// 	OrderClause:           "ORDER BY f.price_next_week ASC",
	// }
	flights, err := ExecuteFlightHistogramQuery(input)
	if err != nil {
		log.Fatalf("ExecuteFlightHistogramQuery failed: %v", err)
	}
	for _, f := range flights {
		fmt.Printf("Destination: %s | Flight Price: %.2f | Active Origin: %s\n", f.DestinationCityName, f.PriceCity1.Float64, f.UrlCity1)
	}
}
