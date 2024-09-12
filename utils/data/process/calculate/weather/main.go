
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Tris20/FairFareFinder/config/handlers"
)


func main() {
	// Check if sufficient arguments are provided
	if len(os.Args) != 4 {
		log.Fatalf("Usage: %s <temperature> <wind_speed> <condition>", os.Args[0])
	}

	// Parse temperature from arguments
	temp, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		log.Fatal("Invalid temperature input:", err)
	}

	// Parse wind speed from arguments
	windSpeed, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		log.Fatal("Invalid wind speed input:", err)
	}

	// Condition is a string, no need to parse
	condition := os.Args[3]

	// Load weather pleasantness config
	config, err := config_handlers.LoadWeatherPleasantnessConfig("../../../../../config/weatherPleasantness.yaml")
	if err != nil {
		log.Fatal("Error loading weather pleasantness config:", err)
	}

	// Calculate the Weather Pleasantness Index (WPI)
	wpi := weatherPleasantness(temp, windSpeed, condition, config)
	fmt.Printf("Weather Pleasantness Index: %.2f\n", wpi)
}

// weatherPleasantness calculates the "weather pleasantness index" (WPI)
func weatherPleasantness(temp float64, wind float64, cond string, config config_handlers.WeatherPleasantnessConfig) float64 {
	weightTemp := 5.0
	weightWind := 1.0
	weightCond := 2.0

	tempIndex := tempPleasantness(temp) * weightTemp
	windIndex := windPleasantness(wind) * weightWind
	weatherIndex := weatherCondPleasantness(cond, config) * weightCond

	index := (tempIndex + windIndex + weatherIndex) / (weightTemp + weightWind + weightCond)
	return index
}

