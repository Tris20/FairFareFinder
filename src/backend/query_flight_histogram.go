package backend

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/Tris20/FairFareFinder/src/backend/config"
	"github.com/Tris20/FairFareFinder/src/backend/model"
	_ "github.com/mattn/go-sqlite3"
)

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
func adjustExpressionForActive(expr Expression, active CityInput, globalThreshold float64) Expression {
	switch e := expr.(type) {
	case *CityCondition:
		if e.City.Name == active.Name && (e.City.Country == active.Country || e.City.Country == "") {
			return &CityCondition{
				City: CityInput{
					Name:       e.City.Name,
					Country:    active.Country, // set explicitly
					PriceLimit: globalThreshold,
				},
			}
		}
		return e
	case *LogicalExpression:
		return &LogicalExpression{
			Operator: e.Operator,
			Left:     adjustExpressionForActive(e.Left, active, globalThreshold),
			Right:    adjustExpressionForActive(e.Right, active, globalThreshold),
		}
	default:
		log.Fatalf("adjustExpressionForActive: unknown expression type")
		return nil
	}
}

func buildQueryForActiveOrigin(active CityInput, cities []string, logicalOperators []string, maxPrices []float64, globalThreshold, accomLimit float64) (string, []interface{}, error) {

	expr, err := ParseLogicalExpression(cities, logicalOperators, maxPrices)
	if err != nil {
		return "", nil, fmt.Errorf("parsing expression: %w", err)
	}

	adjustedExpr := adjustExpressionForActive(expr, active, globalThreshold)

	subquery, subArgs := BuildFlightOriginsSubquery(adjustedExpr)

	withClause := fmt.Sprintf("WITH DestinationSet AS (\n%s\n)", subquery)

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
	finalQuery := withClause + "\n" + mainQuery

	fullArgs := append(subArgs, active.Name, active.Name, active.Country, globalThreshold, accomLimit)

	return finalQuery, fullArgs, nil
}

func ExecuteFlightPricesHistogramQuery(input *FilterInput) ([]model.Flight, error) {

	if db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	cityCountryPairs := GetCityCountryPairs()

	// Derive active origins using input.Cities and input.MaxFlightPrices.
	var activeOrigins []CityInput
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
		activeOrigins = append(activeOrigins, CityInput{
			Name:       city,
			Country:    country,
			PriceLimit: input.MaxFlightPrices[i],
		})
	}

	// Global flight price threshold.
	globalFlightPriceThreshold := 2500.00

	var flights []model.Flight

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
			flight.UrlCity1 = activeOrigin
			flights = append(flights, flight)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("rows error for active origin %s: %w", active.Name, err)
		}
		rows.Close()
	}
	if !config.MutePrints {
		for _, f := range flights {
			fmt.Printf("Destination: %s | Flight Price: %.2f | Active Origin: %s\n", f.DestinationCityName, f.PriceCity1.Float64, f.UrlCity1)
		}
	}
	return flights, nil
}
