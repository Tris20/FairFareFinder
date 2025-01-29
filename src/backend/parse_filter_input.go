package backend

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// FilterInput represents the parsed filter inputs
type FilterInput struct {
	Cities                []string
	LogicalOperators      []string
	MaxFlightPrices       []float64
	MaxAccommodationPrice float64
	OrderClause           string
	LogicalExpression     Expression
}

// parseAndValidateFilterInputs parses and validates filter-related inputs from the HTTP request
func ParseAndValidateFilterInputs(r *http.Request) (*FilterInput, error) {
	cities := r.URL.Query()["city[]"]
	logicalOperators := r.URL.Query()["logical_operator[]"]
	maxFlightPriceLinearStrs := r.URL.Query()["maxFlightPriceLinear[]"]
	maxAccomPriceLinearStrs := r.URL.Query()["maxAccommodationPrice[]"]
	sortOption := r.URL.Query().Get("sort")

	if len(cities) == 0 || len(cities) != len(logicalOperators)+1 || len(cities) != len(maxFlightPriceLinearStrs) {
		return nil, fmt.Errorf("mismatched input lengths. Cities: %d, Operators: %d, Prices: %d",
			len(cities), len(logicalOperators), len(maxFlightPriceLinearStrs))
	}

	maxFlightPrices := make([]float64, 0, len(maxFlightPriceLinearStrs))
	for _, linearStr := range maxFlightPriceLinearStrs {
		linearValue, err := strconv.ParseFloat(linearStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid flight price parameter")
		}
		mappedValue := MapLinearToExponential(linearValue, 50, 1000, 2500)
		maxFlightPrices = append(maxFlightPrices, mappedValue)
	}

	expr, err := parseLogicalExpression(cities, logicalOperators, maxFlightPrices)
	if err != nil {
		return nil, err
	}

	if sortOption == "" {
		sortOption = "best_weather" // default
	}
	orderClause := determineOrderClause(sortOption)

	var maxAccommodationPrice float64
	if len(maxAccomPriceLinearStrs) > 0 {
		accomLinearStr := maxAccomPriceLinearStrs[0]
		accomLinearValue, err := strconv.ParseFloat(accomLinearStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid accommodation price parameter")
		}
		maxAccommodationPrice = MapLinearToExponential(accomLinearValue, 10, 200, 550)
	} else {
		maxAccommodationPrice = 70.0 // Default value
	}

	return &FilterInput{
		Cities:                cities,
		LogicalOperators:      logicalOperators,
		MaxFlightPrices:       maxFlightPrices,
		MaxAccommodationPrice: maxAccommodationPrice,
		OrderClause:           orderClause,
		LogicalExpression:     expr,
	}, nil
}

var orderByClauses = map[string]string{
	"low_price":            "ORDER BY fnf.price_fnaf ASC",
	"high_price":           "ORDER BY fnf.price_fnaf DESC",
	"best_weather":         "ORDER BY avg_wpi DESC",
	"worst_weather":        "ORDER BY avg_wpi ASC",
	"cheapest_hotel":       "ORDER BY a.booking_pppn ASC",
	"most_expensive_hotel": "ORDER BY a.booking_pppn DESC",
	"shortest_flight":      "ORDER BY f.duration_hour_dot_mins ASC",
	"longest_flight":       "ORDER BY f.duration_hour_dot_mins DESC",
}

func determineOrderClause(sortOption string) string {
	if clause, found := orderByClauses[sortOption]; found {
		return clause
	}
	return "ORDER BY avg_wpi DESC" // Default
}

/*---------------Logical Expressions-----------------------*/

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
