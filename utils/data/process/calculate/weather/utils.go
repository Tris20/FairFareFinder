package main

import (
	"github.com/Tris20/FairFareFinder/config/handlers"
)

// Simple math functions and checks go here

func interpolate(temp, temp1, temp2, score1, score2 float64) float64 {
	return ((temp-temp1)/(temp2-temp1))*(score2-score1) + score1
}

// tempPleasantness, windPleasantness, weatherCondPleasantness defined here.

func tempPleasantness(temperature float64) float64 {
	// Optimal range
	if temperature >= 22 && temperature <= 26 {
		return 10
	}

	//     Interpolation between key temperatures below the optimal range
	if temperature > 18 && temperature < 22 {
		return interpolate(temperature, 18, 22, 7, 10)
	}

	// Interpolation between key temperatures above the optimal range
	if temperature > 26 && temperature < 40 {
		return interpolate(temperature, 26, 40, 10, 0)
	}

	// Below 18 down to 0
	if temperature >= 5 && temperature <= 18 {
		return interpolate(temperature, 5, 18, 0, 7)
	}

	// Anything below 0 or above 40
	if temperature <= 5 || temperature >= 40 {
		return 0
	}

	return 0 // Default case if needed
}

// windPleasantness returns a value between 0 and 10 for wind condition pleasantness
func windPleasantness(windSpeed float64) float64 {
	worstWind := 13.8
	if windSpeed >= worstWind {
		return 0
	} else {
		return 10 - windSpeed*10/worstWind
	}
}

// weatherCondPleasantness returns a value between 0 and 10 for weather condition pleasantness
func weatherCondPleasantness(cond string, config config_handlers.WeatherPleasantnessConfig) float64 {
	pleasantness, ok := config.Conditions[cond]
	if !ok {
		return 0
	}
	return pleasantness
}
