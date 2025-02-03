package backend

import (
	"fmt"
	"log"
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

func parseLogicalExpression(cities []string, logicalOperators []string, maxPrices []float64) (Expression, error) {
	// Validate input lengths
	if len(cities) == 0 || len(cities) != len(maxPrices) || len(cities) != len(logicalOperators)+1 {
		return nil, fmt.Errorf("mismatched input lengths")
	}

	// Base case: Only one city
	if len(cities) == 1 {
		log.Printf("parseLogicalExpression: Single city: %s, PriceLimit: %.2f", cities[0], maxPrices[0])
		return &CityCondition{
			City: CityInput{Name: cities[0], PriceLimit: maxPrices[0]},
		}, nil
	}

	// Start with the first city as the base expression
	log.Printf("parseLogicalExpression: Starting with city: %s, PriceLimit: %.2f", cities[0], maxPrices[0])
	var expr Expression = &CityCondition{
		City: CityInput{Name: cities[0], PriceLimit: maxPrices[0]},
	}

	// Process subsequent cities with their logical operators
	for i := 1; i < len(cities); i++ {
		log.Printf("parseLogicalExpression: Adding city: %s, PriceLimit: %.2f with Operator: %s", cities[i], maxPrices[i], logicalOperators[i-1])
		expr = &LogicalExpression{
			Operator: LogicalOperator(logicalOperators[i-1]),
			Left:     expr,
			Right: &CityCondition{
				City: CityInput{Name: cities[i], PriceLimit: maxPrices[i]},
			},
		}
	}

	return expr, nil
}
