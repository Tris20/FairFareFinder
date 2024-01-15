
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// WeatherData represents the structure for OpenWeatherMap API response.
type WeatherData struct {
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Weather []struct {
		Main string `json:"main"`
	} `json:"weather"`
}

// Secrets represents the structure of the secrets.yaml file.
type Secrets struct {
	APIKeys map[string]string `yaml:"api_keys"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Error: No location provided. Please provide a location as a command-line argument.")
	}
	city := strings.Title(os.Args[1]) // Capitalize the first letter of the location argument

	// Load API key from secrets.yaml
	apiKey, err := loadApiKey("ignore/secrets.yaml", "openweathermap.org")
	if err != nil {
		log.Fatal("Error loading API key:", err)
	}

	// Build the API URL with the provided city
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s", city, apiKey)

	// Make the HTTP request
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	// Parse the JSON response
	var weather WeatherData
	if err := json.Unmarshal(body, &weather); err != nil {
		log.Fatalf("Error parsing JSON response: %v", err)
	}

	// Display the weather data
	fmt.Printf("Temperature in %s: %.2f°C\nWind Speed: %.2fm/s\nWeather Condition: %s\nWPI of: %.2f\n",
		city,
		weather.Main.Temp-273.15, // Convert Kelvin to Celsius
		weather.Wind.Speed,
		weather.Weather[0].Main,
        weatherPleasantness(weather.Main.Temp-273.15, weather.Wind.Speed, weather.Weather[0].Main))
}

// loadApiKey loads the API key for a given domain from a YAML file
func loadApiKey(filePath, domain string) (string, error) {
	var secrets Secrets

	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	err = yaml.Unmarshal(yamlFile, &secrets)
	if err != nil {
		return "", err
	}

	apiKey, ok := secrets.APIKeys[domain]
	if !ok {
		return "", fmt.Errorf("API key for %s not found", domain)
	}

	return apiKey, nil
}

// return the "weather pleasentness index" (WPI) between 0 and 10 based on the conditions. (10=best)
func weatherPleasantness(temp float64, wind float64, cond string) float64 {

    // weights for how much the index is affected by a factor compared to others
    weightTemp := 3.0
    weightWind := 1.0
    weightCond := 2.0

    // calculate the average index taking weights into account
    tempindex := tempPleasantness(temp) * weightTemp
    windindex := windPleasantness(wind) * weightWind
    weathindex := weatherCondPleasantness(cond) * weightCond

    index := (tempindex + windindex + weathindex) / (weightTemp + weightWind + weightCond)

    return index
}


// return a value between 0 and 10 how nice the temperature is (10=best)
func tempPleasantness(temperature float64) float64 {

    // configuration of linear slope and cut-off for perfect temp
    GoodTemp := 20.0
    indexAtGoodTemp := 7.0
    PerfectTemp := 23.0
    
    // linear slope is used between 0°C and GoodTemp
    slope := indexAtGoodTemp / GoodTemp

    if temperature <= 0 {
        return 0
    } else if temperature > PerfectTemp {
        return 10
    } else {
        return slope * temperature
    }
}

// return a value between 0 and 10 how nice the weather condition is (10=best)
func weatherCondPleasantness(cond string) float64 {
    
    // weather conditions from openweather.org rated with pleasantness
    weatherConditions := map[string]float64{
        "Thunderstorm": 0,
        "Drizzle":      1,
        "Rain":         0,
        "Snow":         3,
        "Mist":         3,
        "Smoke":        1,
        "Haze":         4,
        "Dust":         2,
        "Fog":          2,
        "Sand":         3,
        "Ash":          1,
        "Squall":       1,
        "Tornado":      0,
        "Clear":        10,
        "Clouds":       7,
    }

    pleasantness, ok := weatherConditions[cond]
    if !ok {
        return 0
    }
    return pleasantness
}

// return a value between 0 and 10 how nice the wind condition is (10=best)
func windPleasantness(windSpeed float64) float64 {
    
    // 0 m/s is perfect, anything higher is linear worse. 13.8 m/s (50 km/h) is worst.
    worstWind := 13.8

    if windSpeed >= worstWind {
        return 0
    } else {
        return 10 - windSpeed*10/worstWind
    }
}

