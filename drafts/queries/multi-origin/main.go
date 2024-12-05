package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// CityInput represents the input for each city
type CityInput struct {
	Name       string
	PriceLimit float64
}

// LogicalOperator represents a logical operator (AND, OR)
type LogicalOperator string

const (
	AndOperator LogicalOperator = "AND"
	OrOperator  LogicalOperator = "OR"
)

// Expression represents a logical expression
type Expression interface{}

// CityCondition represents a condition for a single city
type CityCondition struct {
	City CityInput
}

// LogicalExpression represents a logical combination of expressions
type LogicalExpression struct {
	Operator LogicalOperator
	Left     Expression
	Right    Expression
}

func main() {
	// Example logical expression: (Berlin AND Munich) OR Edinburgh
	expr := &LogicalExpression{
		Operator: AndOperator,
		Left: &LogicalExpression{
			Operator: AndOperator,
			Left:     &CityCondition{City: CityInput{Name: "Berlin", PriceLimit: 200}},
			Right:    &CityCondition{City: CityInput{Name: "Munich", PriceLimit: 200}},
		},
		Right: &CityCondition{City: CityInput{Name: "Edinburgh", PriceLimit: 200}},
	}

	// Build the query
	query, args := buildQuery(expr)

	// Output the query for debugging
	fmt.Println("Generated SQL Query:")
	fmt.Println(query)
	fmt.Println("Arguments:")
	fmt.Println(args)

	// Connect to the SQLite database (replace "your_database.db" with your actual database file)
	db, err := sql.Open("sqlite3", "../../../data/compiled/main.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Execute the query
	rows, err := db.Query(query, args...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	// Process the results
	for rows.Next() {
		var destinationCityName, destinationCountry, originCities string
		var price float64
		// Add other columns as needed

		err = rows.Scan(&destinationCityName, &destinationCountry, &originCities, &price /*, other columns*/)
		if err != nil {
			panic(err)
		}

		// Output the results
		fmt.Printf("Destination: %s, %s\n", destinationCityName, destinationCountry)
		fmt.Printf("Origin Cities: %s\n", originCities)
		fmt.Printf("Price: %.2f\n", price)
		// Output other columns as needed
	}
}

func buildQuery(expr Expression) (string, []interface{}) {
	var queryBuilder strings.Builder
	var args []interface{}

	// Begin the query with the Flights CTE
	queryBuilder.WriteString("WITH DestinationSet AS (\n")

	// Build the subquery based on the logical expression
	subquery, subqueryArgs := buildSubquery(expr)
	queryBuilder.WriteString(subquery)
	queryBuilder.WriteString("\n)")
	args = append(args, subqueryArgs...)

	// Build the rest of the query
	queryBuilder.WriteString(`
    SELECT 
        ds.destination_city_name,
        ds.destination_country,
        GROUP_CONCAT(DISTINCT f.origin_city_name) AS origin_cities,
        MIN(f.price_next_week) AS price
        -- Add other columns as needed
    FROM DestinationSet ds
    JOIN flight f ON ds.destination_city_name = f.destination_city_name 
                   AND ds.destination_country = f.destination_country
    JOIN location l ON ds.destination_city_name = l.city 
                     AND ds.destination_country = l.country
    JOIN weather w ON w.city = ds.destination_city_name 
                    AND w.country = ds.destination_country
    LEFT JOIN accommodation a ON a.city = ds.destination_city_name 
                               AND a.country = ds.destination_country
    WHERE l.avg_wpi BETWEEN 1.0 AND 10.0 
      AND w.date >= date('now')
      AND f.price_next_week < ?
      AND f.origin_city_name IN (?, ?, ?)
    GROUP BY ds.destination_city_name, ds.destination_country
    ORDER BY l.avg_wpi DESC;
    `)

	// Add price limits and origin city names to args
	maxPrice := 2000.0 // Set a max price or calculate based on inputs
	args = append(args, maxPrice)
	args = append(args, "Berlin", "Munich", "Edinburgh")

	return queryBuilder.String(), args
}

func buildSubquery(expr Expression) (string, []interface{}) {
	switch e := expr.(type) {
	case *CityCondition:
		// Return the subquery for a city condition
		subquery := fmt.Sprintf(`
            SELECT 
                f.destination_city_name,
                f.destination_country
            FROM flight f
            WHERE f.origin_city_name = ? AND f.price_next_week < ?
            GROUP BY f.destination_city_name, f.destination_country
        `)
		args := []interface{}{e.City.Name, e.City.PriceLimit}
		return subquery, args
	case *LogicalExpression:
		// Build the left and right subqueries
		leftSubquery, leftArgs := buildSubquery(e.Left)
		rightSubquery, rightArgs := buildSubquery(e.Right)
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
