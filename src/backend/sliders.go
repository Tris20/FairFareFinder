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

	// Get all values of maxPriceLinear[]
	maxPriceLinearStrs := r.URL.Query()["maxPriceLinear[]"]
	if len(maxPriceLinearStrs) == 0 {
		log.Printf("Missing maxPriceLinear parameter")
		http.Error(w, "Missing maxPriceLinear parameter", http.StatusBadRequest)
		return
	}

	// Parse the first value (assuming one slider value per request)
	maxPriceLinearStr := maxPriceLinearStrs[0]
	maxPriceLinear, err := strconv.ParseFloat(maxPriceLinearStr, 64)
	if err != nil {
		log.Printf("Error parsing maxPriceLinear: %v", err)
		http.Error(w, "Invalid maxPrice value", http.StatusBadRequest)
		return
	}

	// Map the slider value to the exponential range
	maxPrice := MapLinearToExponential(maxPriceLinear, 50, 2500)

	// Debug: Log the calculated price
	log.Printf("Calculated price for maxPriceLinear %f: €%.2f", maxPriceLinear, maxPrice)

	// Respond with the formatted price
	fmt.Fprintf(w, "€%.2f", maxPrice)
}
