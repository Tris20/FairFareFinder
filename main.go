// package main

// import (
// 	"fmt"
// )

// func main() {
// 	fmt.Println("Hello World")
// }

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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
	// Load API key from secrets.yaml
	apiKey, err := loadApiKey("ignore/secrets.yaml", "openweathermap.org")
	if err != nil {
		log.Fatal("Error loading API key:", err)
	}

	// Define the city and build the API URL
	city := "Berlin" // Replace with your city name
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
	fmt.Printf("Temperature: %.2f°C\nWind Speed: %.2f \n",
		weather.Main.Temp-273.15, // Convert Kelvin to Celsius
		weather.Wind.Speed)
	// weather.Weather[0].Main)

	// rest of your code...
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
