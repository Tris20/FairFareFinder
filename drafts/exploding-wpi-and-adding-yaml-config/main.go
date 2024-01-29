package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"gopkg.in/yaml.v2"
  "sort"
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

type Favourites struct {
    Locations []string `yaml:"locations"`
}

type ForecastResponse struct {
	List []WeatherData `json:"list"`
}

// Secrets represents the structure of the secrets.yaml file.
type Secrets struct {
	APIKeys map[string]string `yaml:"api_keys"`
}


type CityAverageWPI struct {
    Name string
    WPI  float64
}

func main() {
	if len(os.Args) < 2 {
    log.Fatal("Error: No argument provided. Please provide a location, 'local_favourites', or 'international_favourites'.")
	}
  switch os.Args[1] {
      case "local_favourites":
          handleFavourites("local_favourites.yaml")
      case "international_favourites":
          handleFavourites("international_favourites.yaml")
      default:
          location := strings.Join(os.Args[1:], " ")
          processLocation(location)
  
	}
}

func handleFavourites(yamlFile string)  {
	var favs Favourites
	fileContents, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		log.Fatalf("Error reading favourites file: %v", err)
	}

	err = yaml.Unmarshal(fileContents, &favs)
	if err != nil {
		log.Fatalf("Error parsing favourites file: %v", err)
	}

  var cityWPIs []CityAverageWPI
	for _, location := range favs.Locations {
    wpi := processLocation(location)
    cityWPIs = append(cityWPIs, CityAverageWPI{Name: location, WPI: wpi})
	}
  sort.Slice(cityWPIs, func(i, j int) bool {
        return cityWPIs[i].WPI > cityWPIs[j].WPI
  })

  fmt.Println("\nAverage WPI of Cities (Highest to Lowest):")
  for _, cityWPI := range cityWPIs {
      fmt.Printf("%s: %.2f\n", cityWPI.Name, cityWPI.WPI)
  }
}

func processLocation(location string) float64 {
	// Load API key from secrets.yaml
	apiKey, err := loadApiKey("../../ignore/secrets.yaml", "openweathermap.org")
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
	config, err := LoadWeatherPleasantnessConfig("weatherPleasantness.yaml")
	if err != nil {
		log.Fatal("Error loading weather pleasantness config:", err)
	}
  
  dailyDetails, overallAverage := ProcessForecastData(forecast.List, config)
  displayForecastData(location, dailyDetails)

  return overallAverage
}


func displayForecastData(location string, dailyDetails map[time.Weekday]DailyWeatherDetails) {
    orderedDays := []time.Weekday{time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday, time.Monday, time.Tuesday}

    fmt.Printf("Weather Pleasantness Index (WPI) for %s:\n", location)
    for _, day := range orderedDays {
        details, ok := dailyDetails[day]
        if ok {
            fmt.Printf("%s: Avg Temp: %.2f°C, Weather: %s, WPI: %.2f\n",
                day.String(), details.AverageTemp, details.CommonWeather, details.WPI)
        }
    }
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
