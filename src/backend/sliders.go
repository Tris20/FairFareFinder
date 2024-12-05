package backend

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// UpdateSliderPriceHandler updates the slider price for maxPriceLinear
func UpdateSliderPriceHandler(w http.ResponseWriter, r *http.Request) {
	maxPriceLinearStr := r.URL.Query().Get("maxPriceLinear")
	maxPriceLinear, err := strconv.ParseFloat(maxPriceLinearStr, 64)
	if err != nil {
		log.Printf("Error parsing maxPriceLinear: %v", err)
		http.Error(w, "Invalid maxPrice value", http.StatusBadRequest)
		return
	}

	// Use the function from the mappings file
	maxPrice := MapLinearToExponential(maxPriceLinear, 50, 2500)

	fmt.Fprintf(w, "€%.2f", maxPrice)
}
