package main

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/Tris20/FairFareFinder/src/backend/model"
)

// expectedFlightsGlasgow is the expected output for the Glasgow-only input.
var expectedFlightsGlasgow = []model.Flight{
	{
		DestinationCityName: "Alicante",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Amsterdam",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Belfast",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Berlin",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Birmingham",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Bodrum",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Bristol",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Budapest",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Cork",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Dublin",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Larnaca",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Las Palmas",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Luqa",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Prague",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Saint Helier",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Southampton",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Barcelona",
		PriceCity1:          sql.NullFloat64{Float64: 40.70, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Arrecife",
		PriceCity1:          sql.NullFloat64{Float64: 52.80, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Rome",
		PriceCity1:          sql.NullFloat64{Float64: 83.04, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Marrakech",
		PriceCity1:          sql.NullFloat64{Float64: 96.00, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Split",
		PriceCity1:          sql.NullFloat64{Float64: 100.25, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Faro",
		PriceCity1:          sql.NullFloat64{Float64: 100.72, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Porto",
		PriceCity1:          sql.NullFloat64{Float64: 116.80, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Antalya",
		PriceCity1:          sql.NullFloat64{Float64: 133.18, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Brussels",
		PriceCity1:          sql.NullFloat64{Float64: 138.22, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Derry",
		PriceCity1:          sql.NullFloat64{Float64: 146.74, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Donegal",
		PriceCity1:          sql.NullFloat64{Float64: 188.78, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Ortaca",
		PriceCity1:          sql.NullFloat64{Float64: 206.72, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Burgas",
		PriceCity1:          sql.NullFloat64{Float64: 213.37, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Vienna",
		PriceCity1:          sql.NullFloat64{Float64: 234.18, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Paphos",
		PriceCity1:          sql.NullFloat64{Float64: 267.46, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Dubai",
		PriceCity1:          sql.NullFloat64{Float64: 348.16, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Lerwick",
		PriceCity1:          sql.NullFloat64{Float64: 637.01, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Toronto",
		PriceCity1:          sql.NullFloat64{Float64: 725.49, Valid: true},
		UrlCity1:            "Glasgow",
	},
}

func TestExecuteFlightHistogramQuerySingleGlasgow(t *testing.T) {
	// Create the FilterInput for Glasgow.
	input := &FilterInput{
		Cities:                []string{"Glasgow"},
		LogicalOperators:      []string{}, // Only one city, no operators needed.
		MaxFlightPrices:       []float64{115.00},
		MaxAccommodationPrice: 150.00,
		OrderClause:           "ORDER BY f.price_next_week ASC",
	}

	// Execute the query.
	actualFlights, err := ExecuteFlightHistogramQuery(input)
	if err != nil {
		t.Fatalf("ExecuteFlightHistogramQuery failed: %v", err)
	}

	// For easier comparison, convert both expected and actual flights into a summary struct.
	type flightSummary struct {
		Destination string  // corresponds to model.Flight.DestinationCityName
		Price       float64 // corresponds to model.Flight.PriceCity1.Float64 (rounded to two decimals)
		Active      string  // corresponds to model.Flight.UrlCity1 (active origin)
	}

	var actualSummaries []flightSummary
	for _, f := range actualFlights {
		// Round the price to two decimals.
		price := float64(int(f.PriceCity1.Float64*100+0.5)) / 100
		actualSummaries = append(actualSummaries, flightSummary{
			Destination: f.DestinationCityName,
			Price:       price,
			Active:      f.UrlCity1,
		})
	}

	var expectedSummaries []flightSummary
	for _, f := range expectedFlightsGlasgow {
		expectedSummaries = append(expectedSummaries, flightSummary{
			Destination: f.DestinationCityName,
			Price:       f.PriceCity1.Float64,
			Active:      f.UrlCity1,
		})
	}

	if !reflect.DeepEqual(actualSummaries, expectedSummaries) {
		t.Errorf("Expected flight summaries:\n%v\nGot:\n%v", expectedSummaries, actualSummaries)
	}

	// Optionally log the actual summaries for debugging.
	for i, fs := range actualSummaries {
		t.Logf("Flight %d: Destination: %s | Flight Price: %.2f | Active Origin: %s", i, fs.Destination, fs.Price, fs.Active)
	}
}
