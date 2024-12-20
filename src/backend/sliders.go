package backend

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func UpdateSliderPriceHandler(w http.ResponseWriter, r *http.Request) {
	// Debug: Log the incoming query parameters
	log.Printf("Received request: %v", r.URL.Query())

	// Get flight price slider values
	maxPriceLinearStrs := r.URL.Query()["maxPriceLinear[]"]
	// Get accommodation price slider values
	maxAccomPriceLinearStrs := r.URL.Query()["maxAccommodationPrice[]"]

	var priceType string
	var maxLinearStr string
	var minRange, maxRange float64
	var maxPrice float64

	// Check which type of slider value is provided
	if len(maxPriceLinearStrs) > 0 {
		priceType = "flight"
		maxLinearStr = maxPriceLinearStrs[0]
		minRange = 50
		maxRange = 2500
	} else if len(maxAccomPriceLinearStrs) > 0 {
		priceType = "accommodation"
		maxLinearStr = maxAccomPriceLinearStrs[0]
		minRange = 10
		maxRange = 550
	} else {
		log.Printf("Missing slider parameter (flight or accommodation)")
		http.Error(w, "Missing slider parameter", http.StatusBadRequest)
		return
	}

	// Parse the linear slider value
	maxLinear, err := strconv.ParseFloat(maxLinearStr, 64)
	if err != nil {
		log.Printf("Error parsing %s slider value: %v", priceType, err)
		http.Error(w, fmt.Sprintf("Invalid %s slider value", priceType), http.StatusBadRequest)
		return
	}

	// Map the slider value to the corresponding range
	if priceType == "flight" {
		maxPrice = MapLinearToExponential(maxLinear, minRange, maxRange)
	} else if priceType == "accommodation" {
		maxPrice = AccomMapLinearToExponential(maxLinear, minRange, maxRange)
	}

	// Debug: Log the calculated price
	log.Printf("Calculated %s price for linear value %f: €%.2f", priceType, maxLinear, maxPrice)

	// Respond with the formatted price
	fmt.Fprintf(w, "€%.2f", maxPrice)
}
