package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"

	"github.com/Tris20/FairFareFinder/src/backend"
)

// FlightResult holds one row of the final output.
type FlightResult struct {
	DestinationCity    string
	DestinationCountry string
	FlightPrice        float64
}

// adjustExpressionForActive recursively traverses the Expression tree and, if a CityCondition
// matches the active origin (by Name, and if available, Country), replaces its PriceLimit with globalThreshold.
// If the parsed condition’s Country is empty, we assume it should match.
func adjustExpressionForActive(expr backend.Expression, active backend.CityInput, globalThreshold float64) backend.Expression {
	switch e := expr.(type) {
	case *backend.CityCondition:
		// Check if the city names match and either the country matches or is empty.
		if e.City.Name == active.Name && (e.City.Country == active.Country || e.City.Country == "") {
			// Return a new condition with the adjusted threshold.
			return &backend.CityCondition{
				City: backend.CityInput{
					Name:       e.City.Name,
					Country:    active.Country, // set it explicitly
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
// then adjusts the expression (so that for the active origin the threshold is globalThreshold),
// builds a destination set subquery from that expression, wraps it in a WITH clause,
// and finally joins it with the flight table (restricted to the active origin) plus weather,
// location, and accommodation.
func buildQueryForActiveOrigin(active backend.CityInput, cities []string, logicalOperators []string, maxPrices []float64, globalThreshold, accomLimit float64) (string, []interface{}, error) {
	// 1. Parse the logical expression from the inputs.
	expr, err := backend.ParseLogicalExpression(cities, logicalOperators, maxPrices)
	if err != nil {
		return "", nil, fmt.Errorf("parsing expression: %w", err)
	}

	// 2. Adjust the expression for the active origin.
	adjustedExpr := adjustExpressionForActive(expr, active, globalThreshold)

	// 3. Build the destination set subquery from the adjusted expression.
	subquery, subArgs := backend.BuildFlightOriginsSubquery(adjustedExpr)

	// 4. Wrap the subquery in a WITH clause.
	withClause := fmt.Sprintf("WITH DestinationSet AS (\n%s\n)", subquery)

	// 5. Build the main query.
	mainQuery := `
SELECT DISTINCT 
  ds.destination_city_name,
  ds.destination_country,
  f.price_next_week AS FlightPrice
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
ORDER BY f.price_next_week ASC;
`
	// 6. Combine the WITH clause and main query.
	finalQuery := withClause + "\n" + mainQuery

	// 7. Build the full list of arguments.
	fullArgs := append(subArgs, active.Name, active.Country, globalThreshold, accomLimit)

	return finalQuery, fullArgs, nil
}

func main() {
	// Example inputs:
	//   Cities:      ["Berlin", "Glasgow", "Frankfurt"]
	//   Logical operators between cities: ["AND", "OR"]
	//   MaxPrices:   [1038.00, 1033.00, 1055.00]
	cities := []string{"Berlin", "Glasgow", "Frankfurt"}
	logicalOperators := []string{"AND", "OR"}
	maxPrices := []float64{122.00, 113.00, 96.00}

	// Global values:
	accomPriceLimit := 56.00
	globalFlightPriceThreshold := 2500.00

	// Define active origins.
	activeOrigins := []backend.CityInput{
		{Name: "Berlin", Country: "DE", PriceLimit: 122.00},
		{Name: "Glasgow", Country: "GB", PriceLimit: 113.00},
		{Name: "Frankfurt", Country: "DE", PriceLimit: 96.00},
	}

	// Open the database.
	db, err := sql.Open("sqlite3", "../../data/compiled/main.db")
	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}
	defer db.Close()

	// Map to hold results for each active origin.
	resultsMap := make(map[string][]FlightResult)

	// Loop over each active origin.
	for _, active := range activeOrigins {
		query, args, err := buildQueryForActiveOrigin(active, cities, logicalOperators, maxPrices, globalFlightPriceThreshold, accomPriceLimit)
		if err != nil {
			log.Fatalf("Error building query for active origin %s: %v", active.Name, err)
		}
		log.Printf("Query for active origin %s:\n%s\nArgs: %v", active.Name, query, args)

		rows, err := db.Query(query, args...)
		if err != nil {
			log.Fatalf("DB query failed for active origin %s: %v", active.Name, err)
		}

		var resArr []FlightResult
		for rows.Next() {
			var fr FlightResult
			if err := rows.Scan(&fr.DestinationCity, &fr.DestinationCountry, &fr.FlightPrice); err != nil {
				log.Fatalf("Row scan failed for active origin %s: %v", active.Name, err)
			}
			resArr = append(resArr, fr)
		}
		if err := rows.Err(); err != nil {
			log.Fatalf("Rows error for active origin %s: %v", active.Name, err)
		}
		rows.Close()

		resultsMap[active.Name] = resArr
	}

	// Print the results for each active origin.
	for _, active := range activeOrigins {
		fmt.Printf("Results for active origin %s:\n", active.Name)
		if len(resultsMap[active.Name]) == 0 {
			fmt.Println("  (No results)")
		} else {
			for _, r := range resultsMap[active.Name] {
				fmt.Printf("  Destination: %s, %s | Flight Price: %.2f\n", r.DestinationCity, r.DestinationCountry, r.FlightPrice)
			}
		}
		fmt.Println()
	}
}
