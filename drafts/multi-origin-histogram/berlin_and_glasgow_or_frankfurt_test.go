package main

import (
	"database/sql"
	"testing"

	"github.com/Tris20/FairFareFinder/src/backend/model"
)

// expectedFlights is our expected output from the test.
// Adjust these values to exactly match the expected output for your test data.
var expectedFlights_and_or = []model.Flight{
	{
		DestinationCityName: "Luqa",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Berlin",
	},
	{
		DestinationCityName: "Split",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
		UrlCity1:            "Berlin",
	},
	{
		DestinationCityName: "Faro",
		PriceCity1:          sql.NullFloat64{Float64: 116.19, Valid: true},
		UrlCity1:            "Berlin",
	},
	{
		DestinationCityName: "Larnaca",
		PriceCity1:          sql.NullFloat64{Float64: 162.89, Valid: true},
		UrlCity1:            "Berlin",
	},
	{
		DestinationCityName: "Luqa",
		PriceCity1:          sql.NullFloat64{Float64: 0.00, Valid: true},
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
		DestinationCityName: "Antalya",
		PriceCity1:          sql.NullFloat64{Float64: 133.18, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Paphos",
		PriceCity1:          sql.NullFloat64{Float64: 267.46, Valid: true},
		UrlCity1:            "Glasgow",
	},
	{
		DestinationCityName: "Ankara",
		PriceCity1:          sql.NullFloat64{Float64: 115.05, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Antalya",
		PriceCity1:          sql.NullFloat64{Float64: 115.16, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Agadir",
		PriceCity1:          sql.NullFloat64{Float64: 128.89, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Tirana",
		PriceCity1:          sql.NullFloat64{Float64: 152.65, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Sofia",
		PriceCity1:          sql.NullFloat64{Float64: 162.00, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Tunis",
		PriceCity1:          sql.NullFloat64{Float64: 170.00, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Skopje",
		PriceCity1:          sql.NullFloat64{Float64: 174.16, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Bucharest",
		PriceCity1:          sql.NullFloat64{Float64: 178.69, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Sarajevo",
		PriceCity1:          sql.NullFloat64{Float64: 187.87, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Larnaca",
		PriceCity1:          sql.NullFloat64{Float64: 195.24, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Tallinn",
		PriceCity1:          sql.NullFloat64{Float64: 199.84, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Casablanca",
		PriceCity1:          sql.NullFloat64{Float64: 201.80, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Marseille",
		PriceCity1:          sql.NullFloat64{Float64: 208.59, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Zagreb",
		PriceCity1:          sql.NullFloat64{Float64: 214.53, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Yerevan",
		PriceCity1:          sql.NullFloat64{Float64: 220.55, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Faro",
		PriceCity1:          sql.NullFloat64{Float64: 229.57, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Vilnius",
		PriceCity1:          sql.NullFloat64{Float64: 235.06, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Riga",
		PriceCity1:          sql.NullFloat64{Float64: 235.95, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Baku",
		PriceCity1:          sql.NullFloat64{Float64: 259.38, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Catania",
		PriceCity1:          sql.NullFloat64{Float64: 261.34, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Warsaw",
		PriceCity1:          sql.NullFloat64{Float64: 265.54, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Muscat",
		PriceCity1:          sql.NullFloat64{Float64: 269.56, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Katowice",
		PriceCity1:          sql.NullFloat64{Float64: 276.03, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Beirut",
		PriceCity1:          sql.NullFloat64{Float64: 277.02, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Luqa",
		PriceCity1:          sql.NullFloat64{Float64: 277.70, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Nantes",
		PriceCity1:          sql.NullFloat64{Float64: 281.83, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Amman",
		PriceCity1:          sql.NullFloat64{Float64: 307.90, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Split",
		PriceCity1:          sql.NullFloat64{Float64: 326.30, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Astana",
		PriceCity1:          sql.NullFloat64{Float64: 371.60, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Graz",
		PriceCity1:          sql.NullFloat64{Float64: 391.12, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Almaty",
		PriceCity1:          sql.NullFloat64{Float64: 416.00, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Tashkent",
		PriceCity1:          sql.NullFloat64{Float64: 500.04, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "New Delhi",
		PriceCity1:          sql.NullFloat64{Float64: 517.15, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Addis Ababa",
		PriceCity1:          sql.NullFloat64{Float64: 519.45, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Calgary",
		PriceCity1:          sql.NullFloat64{Float64: 606.76, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Phuket",
		PriceCity1:          sql.NullFloat64{Float64: 666.98, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Port Louis",
		PriceCity1:          sql.NullFloat64{Float64: 754.84, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Mombasa",
		PriceCity1:          sql.NullFloat64{Float64: 759.02, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Nairobi",
		PriceCity1:          sql.NullFloat64{Float64: 774.53, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Johannesburg",
		PriceCity1:          sql.NullFloat64{Float64: 782.11, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Atlanta",
		PriceCity1:          sql.NullFloat64{Float64: 788.05, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Hanoi",
		PriceCity1:          sql.NullFloat64{Float64: 791.06, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Houston",
		PriceCity1:          sql.NullFloat64{Float64: 791.17, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Colombo",
		PriceCity1:          sql.NullFloat64{Float64: 825.88, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Varadero",
		PriceCity1:          sql.NullFloat64{Float64: 867.46, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Abuja",
		PriceCity1:          sql.NullFloat64{Float64: 895.94, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "Windhoek",
		PriceCity1:          sql.NullFloat64{Float64: 992.09, Valid: true},
		UrlCity1:            "Frankfurt",
	},
	{
		DestinationCityName: "La Romana",
		PriceCity1:          sql.NullFloat64{Float64: 1225.54, Valid: true},
		UrlCity1:            "Frankfurt",
	},
}

func TestExecuteFlightHistogramQuery_and_or(t *testing.T) {
	input := &FilterInput{
		Cities:                []string{"Berlin", "Glasgow", "Frankfurt"},
		LogicalOperators:      []string{"AND", "OR"},
		MaxFlightPrices:       []float64{122.00, 113.00, 96.00},
		MaxAccommodationPrice: 56.00,
		OrderClause:           "ORDER BY f.price_next_week ASC",
	}

	actualFlights, err := ExecuteFlightHistogramQuery(input)
	if err != nil {
		t.Fatalf("ExecuteFlightHistogramQuery_and_or failed: %v", err)
	}

	// Compare lengths.
	if len(actualFlights) != len(expectedFlights_and_or) {
		t.Errorf("Expected %d flights, got %d", len(expectedFlights_and_or), len(actualFlights))
	}

	// Compare each flight.
	for i, exp := range expectedFlights_and_or {
		if i >= len(actualFlights) {
			break
		}
		act := actualFlights[i]
		if exp.DestinationCityName != act.DestinationCityName ||
			!exp.PriceCity1.Valid || !act.PriceCity1.Valid ||
			exp.PriceCity1.Float64 != act.PriceCity1.Float64 ||
			exp.UrlCity1 != act.UrlCity1 {
			t.Errorf("Flight %d mismatch.\nExpected: %#v\nGot: %#v", i, exp, act)
		}
	}
}
