package main

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/Tris20/FairFareFinder/src/backend/model"
)

// expectedFlights is the expected output for the test.
var expectedFlights_and_and = []model.Flight{
	{
		DestinationCityName: "Alicante",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Berlin",
	},
	{
		DestinationCityName: "Dublin",
		PriceCity1:          sql.NullFloat64{Float64: 19.98, Valid: true},
		UrlCity1:            "Berlin",
	},
	{
		DestinationCityName: "Arrecife",
		PriceCity1:          sql.NullFloat64{Float64: 76.01, Valid: true},
		UrlCity1:            "Berlin",
	},
	{
		DestinationCityName: "Las Palmas",
		PriceCity1:          sql.NullFloat64{Float64: 116.05, Valid: true},
		UrlCity1:            "Berlin",
	},
	{
		DestinationCityName: "Alicante",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Dublin",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Arrecife",
		PriceCity1:          sql.NullFloat64{Float64: 52.80, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Brussels",
		PriceCity1:          sql.NullFloat64{Float64: 138.22, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Dublin",
		PriceCity1:          sql.NullFloat64{Float64: 114.53, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Arrecife",
		PriceCity1:          sql.NullFloat64{Float64: 122.16, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Alicante",
		PriceCity1:          sql.NullFloat64{Float64: 168.12, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Barcelona",
		PriceCity1:          sql.NullFloat64{Float64: 184.45, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Budapest",
		PriceCity1:          sql.NullFloat64{Float64: 198.39, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Birmingham",
		PriceCity1:          sql.NullFloat64{Float64: 208.85, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Rome",
		PriceCity1:          sql.NullFloat64{Float64: 210.43, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Marrakech",
		PriceCity1:          sql.NullFloat64{Float64: 222.08, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Luqa",
		PriceCity1:          sql.NullFloat64{Float64: 277.70, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Split",
		PriceCity1:          sql.NullFloat64{Float64: 326.30, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	// (Add additional expected rows as needed.)
}

// TestExecuteFlightHistogramQuery tests the ExecuteFlightHistogramQuery function.
func TestExecuteFlightHistogramQuery_and_and(t *testing.T) {
	// Create the FilterInput.
	input := &FilterInput{
		Cities:                []string{"Berlin", "Glasgow", "Frankfurt"},
		LogicalOperators:      []string{"AND", "AND"},
		MaxFlightPrices:       []float64{115.00, 107.00, 177.00},
		MaxAccommodationPrice: 150.00,
		OrderClause:           "ORDER BY f.price_next_week ASC",
	}

	// Execute the query.
	actualFlights, err := ExecuteFlightHistogramQuery(input)
	if err != nil {
		t.Fatalf("ExecuteFlightHistogramQuery_and_and failed: %v", err)
	}

	// For easier comparison, you might want to filter out any fields you don't care about.
	// Here we compare DestinationCityName, PriceCity1, and UrlCity1.
	type flightSummary struct {
		Destination string
		Price       float64
		Active      string
	}

	var actualSummaries []flightSummary
	for _, f := range actualFlights {
		// Round the price to two decimals (if needed)
		price := f.PriceCity1.Float64
		actualSummaries = append(actualSummaries, flightSummary{
			Destination: f.DestinationCityName,
			Price:       price,
			Active:      f.UrlCity1,
		})
	}

	var expectedSummaries []flightSummary
	for _, f := range expectedFlights_and_and {
		expectedSummaries = append(expectedSummaries, flightSummary{
			Destination: f.DestinationCityName,
			Price:       f.PriceCity1.Float64,
			Active:      f.UrlCity1,
		})
	}

	// Compare the slices.
	if !reflect.DeepEqual(actualSummaries, expectedSummaries) {
		t.Errorf("Expected flights:\n%v\nGot:\n%v", expectedSummaries, actualSummaries)
	}

	// Optionally, print out the actual summaries for debugging.
	for i, fs := range actualSummaries {
		t.Logf("Flight %d: Destination: %s | Flight Price: %.2f | Active Origin: %s", i, fs.Destination, fs.Price, fs.Active)
	}
}
