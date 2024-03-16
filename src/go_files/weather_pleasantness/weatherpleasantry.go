package weather_pleasantry

import (
	"encoding/json"
	"fmt"
	"github.com/Tris20/FairFareFinder/src/go_files"
	"github.com/Tris20/FairFareFinder/src/go_files/config_handlers"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type ForecastResponse struct {
	List []model.WeatherData `json:"list"`
}

func ProcessLocation(location string) (float64, map[time.Weekday]model.DailyWeatherDetails) {

	// Load API key from secrets.yaml
	apiKey, err := config_handlers.LoadApiKey("ignore/secrets.yaml", "openweathermap.org")
	if err != nil {
		log.Fatal("Error loading API key:", err)
	}

	// Build the forecast API URL with the provided city
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/forecast?q=%s&appid=%s&units=metric", location, apiKey)

	// Make the HTTP request
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response body
	var forecast ForecastResponse
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		log.Fatalf("Error reading response body: %v", err)
	} else if err := json.Unmarshal(body, &forecast); err != nil {
		log.Fatalf("Error parsing JSON response: %v", err)
	}

	// Load weather pleasantness config
	config, err := LoadWeatherPleasantnessConfig("input/weatherPleasantness.yaml")
	if err != nil {
		log.Fatal("Error loading weather pleasantness config:", err)
	}

	dailyDetails, overallAverage := ProcessForecastData(forecast.List, config)
	DisplayForecastData(location, dailyDetails)

	return overallAverage, dailyDetails
}

func DisplayForecastData(location string, dailyDetails map[time.Weekday]model.DailyWeatherDetails) {
	orderedDays := []time.Weekday{time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday, time.Monday, time.Tuesday}

	fmt.Printf("Weather Pleasantness Index (WPI) for %s:\n", location)
	for _, day := range orderedDays {
		details, ok := dailyDetails[day]
		wind_kmh := 3.6 * details.AverageWind
		if ok {
			fmt.Printf("%s: Avg Temp: %.2f°C, Weather: %s, Wind: %.2fkm/h, WPI: %.2f\n",
				day.String(), details.AverageTemp, details.CommonWeather, wind_kmh, details.WPI)
		}
	}
}
