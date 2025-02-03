package backend

import (
	"fmt"
	"github.com/Tris20/FairFareFinder/src/backend/config"
	"github.com/Tris20/FairFareFinder/src/backend/model"
	"log"
	"strings"
)

func ExecuteMainQuery(input *FilterInput) ([]model.Flight, error) {

	query, args := BuildMainQuery(input.LogicalExpression, input.MaxAccommodationPrice, input.Cities, input.OrderClause)

	if !config.MutePrints {
		fmt.Println("Generated SQL Query (MAIN):")
		fmt.Println(query)
		fmt.Println("Arguments:", args)
	}

	fullQuery, err := ReplacePlaceholdersWithArgs(query, args)

	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(fullQuery)
	}

	log.Printf("Full MAIN Query:\n%s\n", fullQuery)

	if db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error querying main results: %v", err)
		return nil, err
	}
	defer rows.Close()

	flights, err := ProcessFlightRows(rows)
	if err != nil {
		log.Printf("Error processing flight rows: %v", err)
		return nil, err
	}

	return flights, nil
}

/*
#
#
#
*/

// Unified Query Builder

func BuildMainQuery(expr Expression, maxAccommodationPrice float64, originCities []string, orderClause string) (string, []interface{}) {
	var queryBuilder strings.Builder
	var args []interface{}
	// Begin the query with the DestinationSet CTE
	queryBuilder.WriteString("WITH DestinationSet AS (\n")

	// Build the subquery based on the logical expression
	subquery, subqueryArgs := BuildFlightOriginsSubquery(expr)
	queryBuilder.WriteString(subquery)
	queryBuilder.WriteString("\n)")
	args = append(args, subqueryArgs...)

	// This is where the core part of the sql query comes from
	queryBuilder.WriteString(BaseQuery)

	// Build the IN clause dynamically based on the number of origin cities
	if len(originCities) == 0 {
		return "", nil // If no origin cities, return an empty query
	}

	placeholders := make([]string, len(originCities))
	for i := range originCities {
		placeholders[i] = "?"
	}
	inClause := fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))
	queryBuilder.WriteString(inClause)
	queryBuilder.WriteString(`
      AND a.booking_pppn IS NOT NULL
      AND a.booking_pppn <= ?
   GROUP BY f.destination_city_name, w.date, f.destination_country, l.avg_wpi
    `)

	// Add the dynamic order clause
	queryBuilder.WriteString(orderClause)
	queryBuilder.WriteString(";")

	// Add price limits and origin city names to args
	maxPrice := 2000.0 // Set a max price or calculate based on inputs
	args = append(args, maxPrice)

	// Correctly append originCities to args
	for _, city := range originCities {
		args = append(args, city)
	}

	args = append(args, maxAccommodationPrice)

	return queryBuilder.String(), args
}

// Maps to "city-rows"
func BuildFlightOriginsSubquery(expr Expression) (string, []interface{}) {
	switch e := expr.(type) {
	case *CityCondition:
		// Return the subquery for a city condition
		subquery := `
            SELECT 
                f.destination_city_name,
                f.destination_country
            FROM flight f
            WHERE f.origin_city_name = ? AND f.price_next_week < ?
            GROUP BY f.destination_city_name, f.destination_country
        `
		args := []interface{}{e.City.Name, e.City.PriceLimit}
		return subquery, args
	case *LogicalExpression:
		// Build the left and right subqueries
		leftSubquery, leftArgs := BuildFlightOriginsSubquery(e.Left)
		rightSubquery, rightArgs := BuildFlightOriginsSubquery(e.Right)
		var operator string
		if e.Operator == AndOperator {
			operator = "INTERSECT"
		} else if e.Operator == OrOperator {
			operator = "UNION"
		} else {
			panic("Unknown operator")
		}
		// Combine subqueries without unnecessary parentheses
		combinedSubquery := fmt.Sprintf("%s\n%s\n%s", leftSubquery, operator, rightSubquery)
		args := append(leftArgs, rightArgs...)
		return combinedSubquery, args
	default:
		panic("Unknown expression type")
	}
}
