package backend

import (
	"database/sql"
	"math"
)

// Generalized function to map a linear slider (0-100) to an exponential range
func MapLinearToExponential(linearValue float64, minVal float64, midVal float64, maxVal float64) float64 {
	percentage := linearValue / 100

	// First 70% of the slider covers minVal to midVal
	if percentage <= 0.7 {
		return minVal * math.Pow(midVal/minVal, percentage/0.7)
	} else {
		// Last 30% covers midVal to maxVal
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
