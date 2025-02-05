package backend

import (
	"fmt"
	"github.com/Tris20/FairFareFinder/src/backend/config"
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
		mappedValue := MapLinearToExponential(linearValue, config.MinFlightPrice, config.MidFlightPrice, config.MaxFlightPrice)
		maxFlightPrices = append(maxFlightPrices, mappedValue)
	}

	expr, err := ParseLogicalExpression(cities, logicalOperators, maxFlightPrices)
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
		maxAccommodationPrice = MapLinearToExponential(accomLinearValue, config.MinAccomPrice, config.MidAccomPrice, config.MaxAccomPrice)
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
