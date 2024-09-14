
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
"io/ioutil"
	"github.com/Tris20/FairFareFinder/config/handlers"
)

// WeatherData represents the structure of weather information to be stored in weather.db
type WeatherData struct {
	Date              string
	WeatherType       string
	Temperature       float64
	WeatherIconURL    string
	GoogleWeatherLink string
	WindSpeed         float64 // New field for wind speed
}

// Assuming a part of the JSON response structure from OpenWeatherMap API for simplification
type ApiResponse struct {
	List []struct {
		Dt   int64 `json:"dt"`
		Main struct {
			Temp float64 `json:"temp"`
		} `json:"main"`
		Weather []struct {
			Main string `json:"main"`
			Icon string `json:"icon"`
		} `json:"weather"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
	} `json:"list"`
}

// fetchWeatherForCity fetches weather data for the specified city from OpenWeatherAPI
func fetchWeatherForCity(cityName string, countryCode string) ([]WeatherData, error) {
	// Placeholder for OpenWeatherAPI request. Assume you replace the following URL with the actual API request
	location_string := url.QueryEscape(fmt.Sprintf("%s, %s", cityName, countryCode))
	apiKey, err := config_handlers.LoadApiKey("../../../../ignore/secrets.yaml", "openweathermap.org")
	if err != nil {
		return nil, err
	}
	apiURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?q=%s&appid=%s&units=metric", location_string, apiKey)

  fmt.Printf("\napi url: %s \n", apiURL)

/*
  resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
*/



req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()


if resp.StatusCode != 200 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		return nil, fmt.Errorf("received non-200 status code from weather API: %d, response: %s", resp.StatusCode, bodyString)
	}  else{
   fmt.Printf("\nWeather Data for %s %scollected successfully\n",cityName,countryCode)
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
		windSpeed := item.Wind.Speed // Extract wind speed

		googleWeatherURL := fmt.Sprintf("https://www.google.com/search?q=weather+%s", location_string)

		result = append(result, WeatherData{
			Date:              date,
			WeatherType:       weatherType,
			Temperature:       temp,
			WeatherIconURL:    iconURL,
			GoogleWeatherLink: googleWeatherURL,
			WindSpeed:         windSpeed, // Include wind speed
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
		err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM all_weather WHERE city_name = ? AND country_code = ? AND iata = ? AND date = ?)`,
			airport.City, airport.Country, airport.IATA, wd.Date).Scan(&exists)
		if err != nil {
			log.Printf("Failed to check existence for %s on %s: %v", airport.City, wd.Date, err)
			continue
		}

		if exists {
			// If it exists, update the entry
			_, err = db.Exec(`UPDATE all_weather SET weather_type = ?, temperature = ?, weather_icon_url = ?, google_weather_link = ?, wind_speed = ? WHERE city_name = ? AND country_code = ? AND iata = ? AND date = ?`,
				wd.WeatherType, wd.Temperature, wd.WeatherIconURL, wd.GoogleWeatherLink, wd.WindSpeed, airport.City, airport.Country, airport.IATA, wd.Date)
		} else {
			// If it does not exist, insert a new entry
			_, err = db.Exec(`INSERT INTO all_weather (city_name, country_code, iata, date, weather_type, temperature, weather_icon_url, google_weather_link, wind_speed) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				airport.City, airport.Country, airport.IATA, wd.Date, wd.WeatherType, wd.Temperature, wd.WeatherIconURL, wd.GoogleWeatherLink, wd.WindSpeed)
		}

		if err != nil {
			log.Printf("Failed to insert/update weather data for %s: %v", airport.City, err)
			// Decide how to handle the error; continue with the next iteration or return the error
		}
	}
	return nil
}

