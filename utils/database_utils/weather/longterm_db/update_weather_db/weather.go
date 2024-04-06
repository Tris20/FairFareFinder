
package main

import (
  "encoding/json"
  "database/sql"
  "fmt"
	"log"
	"net/http"
	"time"
 "github.com/Tris20/FairFareFinder/src/go_files/config_handlers"
  "net/url"

)

// WeatherData represents the structure of weather information to be stored in weather.db
type WeatherData struct {
	Date            string
	WeatherType     string
	Temperature     float64
	WeatherIconURL  string
	GoogleWeatherLink string
}

// Assuming a part of the JSON response structure from OpenWeatherMap API for simplification
type ApiResponse struct {
	List []struct {
		Dt  int64 `json:"dt"`
		Main struct {
			Temp float64 `json:"temp"`
		} `json:"main"`
		Weather []struct {
			Main        string `json:"main"`
			Icon        string `json:"icon"`
		} `json:"weather"`
	} `json:"list"`
}

// fetchWeatherForCity fetches weather data for the specified city from OpenWeatherAPI
func fetchWeatherForCity(cityName string, countryCode string) ([]WeatherData, error) {
	// Placeholder for OpenWeatherAPI request. Assume you replace the following URL with the actual API request
	location_string := url.QueryEscape(fmt.Sprintf("%s, %s", cityName, countryCode))
  apiKey, err := config_handlers.LoadApiKey("../../../../../ignore/secrets.yaml", "openweathermap.org")
	apiURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?q=%s&appid=%s&units=metric", location_string, apiKey)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("received non-200 status code from weather API")
	}

	var apiResp ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	var result []WeatherData // Assume proper JSON decoding based on OpenWeatherAPI response structure

	for _, item := range apiResp.List {
		// This example extracts weather data for each time entry in the list.
		// You might want to adjust this to extract daily averages or specific times of day.
		date := time.Unix(item.Dt, 0).Format("2006-01-02 15:04:05")
		temp := item.Main.Temp
		weatherType := "Clear" // Default to clear, adjust based on actual data
		if len(item.Weather) > 0 {
			weatherType = item.Weather[0].Main
		}
		iconURL := fmt.Sprintf("https://openweathermap.org/img/w/%s.png", item.Weather[0].Icon)

		googleWeatherURL := fmt.Sprintf("https://www.google.com/search?q=weather+%s", location_string)
		
    result = append(result, WeatherData{
			Date:            date,
			WeatherType:     weatherType,
			Temperature:     temp,
			WeatherIconURL:  iconURL,
			GoogleWeatherLink: googleWeatherURL, 
    })
	}

	return result, nil
}
// storeWeatherData stores the fetched weather data into the weather.db
func storeWeatherData(dbPath string, airport AirportInfo, weatherData []WeatherData) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	 for _, wd := range weatherData {
        // First, check if an entry exists
        var exists bool
        err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM Weather WHERE CityName = ? AND CountryCode = ? AND Date = ?)`,
            airport.City, airport.Country, wd.Date).Scan(&exists)
        if err != nil {
            log.Printf("Failed to check existence for %s on %s: %v", airport.City, wd.Date, err)
            continue
        }

        if exists {
            // If it exists, update the entry
            _, err = db.Exec(`UPDATE Weather SET WeatherType = ?, Temperature = ?, WeatherIconURL = ?, GoogleWeatherLink = ? WHERE CityName = ? AND CountryCode = ? AND Date = ?`,
                wd.WeatherType, wd.Temperature, wd.WeatherIconURL, wd.GoogleWeatherLink, airport.City, airport.Country, wd.Date)
        } else {
            // If it does not exist, insert a new entry
            _, err = db.Exec(`INSERT INTO Weather (CityName, CountryCode, Date, WeatherType, Temperature, WeatherIconURL, GoogleWeatherLink) VALUES (?, ?, ?, ?, ?, ?, ?)`,
                airport.City, airport.Country, wd.Date, wd.WeatherType, wd.Temperature, wd.WeatherIconURL, wd.GoogleWeatherLink)
        }

        if err != nil {
            log.Printf("Failed to insert/update weather data for %s: %v", airport.City, err)
            // Decide how to handle the error; continue with the next iteration or return the error
        }
    }
	return nil
}

