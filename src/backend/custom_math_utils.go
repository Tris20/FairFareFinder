
package backend

import (
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
