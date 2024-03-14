package main

import (
	"github.com/Tris20/FairFareFinder/src/go_files/weather_pleasantness"
	"github.com/Tris20/FairFareFinder/src/go_files"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

// test for displayForecastData function
func TestDisplayForecastData(t *testing.T) {
	location := "New York"
	dailyDetails := map[time.Weekday]model.DailyWeatherDetails{
		time.Wednesday: {
			AverageTemp:   25.0,
			CommonWeather: "Sunny",
			AverageWind:   10.0,
			WPI:           8.5,
		},
		time.Thursday: {
			AverageTemp:   22.5,
			CommonWeather: "Cloudy",
			AverageWind:   15.0,
			WPI:           7.2,
		},
		time.Friday: {
			AverageTemp:   20.0,
			CommonWeather: "Rainy",
			AverageWind:   12.5,
			WPI:           6.0,
		},
	}

	expectedOutput := `Weather Pleasantness Index (WPI) for New York:
Wednesday: Avg Temp: 25.00°C, Weather: Sunny, Wind: 36.00km/h, WPI: 8.50
Thursday: Avg Temp: 22.50°C, Weather: Cloudy, Wind: 54.00km/h, WPI: 7.20
Friday: Avg Temp: 20.00°C, Weather: Rainy, Wind: 45.00km/h, WPI: 6.00
`

	// Redirect stdout to capture the output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	weather_pleasantry.DisplayForecastData(location, dailyDetails)

	// Reset stdout
	w.Close()
	os.Stdout = old

	out, _ := ioutil.ReadAll(r)
	actualOutput := string(out)

	if actualOutput != expectedOutput {
		t.Errorf("Expected output:\n%s\n\nActual output:\n%s", expectedOutput, actualOutput)
	}
}
