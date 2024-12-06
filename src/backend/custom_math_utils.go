package backend

import (
	"database/sql"
	"math"
)

// This function maps a linear slider (0-100) to an exponential range (10-2500)
func MapLinearToExponential(linearValue float64, minVal float64, maxVal float64) float64 {
	midVal := 1000.0
	percentage := linearValue / 100

	// First 70% of the slider covers 10 to 1000
	if percentage <= 0.7 {
		return minVal * math.Pow(midVal/minVal, percentage/0.7)
	} else {
		// Last 30% covers 1000 to 2500
		newPercentage := (percentage - 0.7) / 0.3
		return midVal + (maxVal-midVal)*newPercentage
	}
}

// Helper function to update max value
func UpdateMaxValue(currentMax, newValue sql.NullFloat64) sql.NullFloat64 {
	if !currentMax.Valid || (newValue.Valid && newValue.Float64 > currentMax.Float64) {
		return newValue
	}
	return currentMax
}

// Helper function to update min value
func UpdateMinValue(currentMin, newValue sql.NullFloat64) sql.NullFloat64 {
	// HOTFIX Check if newValue is valid and greater than or equal to 0.1
	// This ensures we don't include flight prices which are zero because no price was found
	if newValue.Valid && newValue.Float64 >= 0.1 {
		// Update currentMin if it's not valid or if newValue is smaller
		if !currentMin.Valid || newValue.Float64 < currentMin.Float64 {
			return newValue
		}
	}
	// Return currentMin if none of the above conditions are met
	return currentMin
}
