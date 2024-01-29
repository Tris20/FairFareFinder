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

type WeatherData struct {
	Dt     int64 `json:"dt"` // Unix timestamp of the forecasted data
	Main   struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Wind   struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Weather []struct {
		Main string `json:"main"`
	} `json:"weather"`
}

type ForecastResponse struct {
	List []WeatherData `json:"list"`
}

// Secrets represents the structure of the secrets.yaml file.
type Secrets struct {
	APIKeys map[string]string `yaml:"api_keys"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Error: No location provided. Please provide a location as a command-line argument.")
	}
	city := strings.Title(os.Args[1])

	// Load API key from secrets.yaml
	apiKey, err := loadApiKey("../../ignore/secrets.yaml", "openweathermap.org")
	if err != nil {
		log.Fatal("Error loading API key:", err)
	}

	// Build the forecast API URL with the provided city
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/forecast?q=%s&appid=%s&units=metric", city, apiKey)

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
  
  // After reading the response body
//  fmt.Println(string(body)) // Print raw JSON response


  // Parse the JSON forecast response
 	var forecast ForecastResponse
	if err := json.Unmarshal(body, &forecast); err != nil {
		log.Fatalf("Error parsing JSON response: %v", err)
	}

	// Load weather pleasantness config
	config, err := LoadWeatherPleasantnessConfig("weatherPleasantness.yaml")
	if err != nil {
		log.Fatal("Error loading weather pleasantness config:", err)
	}

	// Process the forecast data
	dailyAverages, overallAverage := ProcessForecastData(forecast.List, config)

	// Display the results
	fmt.Printf("Weather Pleasantness Index (WPI) for %s:\n", city)
	for day, avgWPI := range dailyAverages {
		fmt.Printf("%s: %.2f\n", day.String(), avgWPI)
	}
	fmt.Printf("Average WPI (Thursday to Sunday): %.2f\n", overallAverage)
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
