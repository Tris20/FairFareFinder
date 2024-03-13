package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// Define structs to unmarshal JSON data
type WeatherResponse struct {
	List []struct {
		Dt   int64 `json:"dt"`
		Main struct {
			Temp float64 `json:"temp"`
		} `json:"main"`
		Weather []struct {
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
	} `json:"list"`
}

// Define a struct for template data
type ForecastDay struct {
	Date        string
	Temperature string
	Weather     string
	IconURL     string
}

func fetchWeatherData() ([]ForecastDay, error) {
	// Replace "your_api_key_here" with your actual OpenWeather API key
	apiKey, err := LoadApiKey("../../ignore/secrets.yaml", "openweathermap.org")
	fmt.Println(apiKey)
	str_get := fmt.Sprintf("http://api.openweathermap.org/data/2.5/forecast?q=Berlin&units=metric&appid=%s", apiKey)
	resp, err := http.Get(str_get)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var weatherResponse WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResponse); err != nil {
		return nil, err
	}
	// Add a debug log to print the weather response or part of it to verify
	log.Println("API response:", weatherResponse)
	var forecastDays []ForecastDay
	forecastCount := len(weatherResponse.List)
	if forecastCount > 4 {
		forecastCount = 4 // Ensure we only take the first 4 forecasts if available
	}

	for _, item := range weatherResponse.List[:forecastCount] {
		//ensure we get the daytime icon by replacing n with d

		iconCode := item.Weather[0].Icon                      // Original icon code, e.g., "10n"
		iconCodeDay := strings.Replace(iconCode, "n", "d", 1) // Replace "n" with "d"
		iconURL := fmt.Sprintf("http://openweathermap.org/img/wn/%s.png", iconCodeDay)
		forecastDays = append(forecastDays, ForecastDay{
			Date:        time.Unix(item.Dt, 0).Format("2006-01-02 15:04:05"),
			Temperature: fmt.Sprintf("%.2f°C", item.Main.Temp),
			Weather:     item.Weather[0].Description,
			IconURL:     iconURL,
		})
	}

	log.Println("Forecast days:", forecastDays)
	return forecastDays, nil
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		forecastDays, err := fetchWeatherData()
		if err != nil {
			http.Error(w, "Failed to fetch weather data", http.StatusInternalServerError)
			log.Println("Error fetching weather data:", err)
			return
		}

		tmpl := template.Must(template.New("forecast").Parse(`
<html>
<head>
    <title>Weather Forecast</title>
</head>
<body>
    <h1>4-Day Weather Forecast</h1>
    <table>
        <tr>
            <th>Date and Time</th>
            <th>Temperature</th>
            <th>Weather</th>
            <th>Icon</th>
        </tr>
        {{range .}}
        <tr>
            <td>{{.Date}}</td>
            <td>{{.Temperature}}</td>
            <td>{{.Weather}}</td>
            <td><img src="{{.IconURL}}" alt="Weather icon"></td>
        </tr>
        {{end}}
    </table>
</body>
</html>
`))

		if err := tmpl.Execute(w, forecastDays); err != nil {
			http.Error(w, "Failed to execute template", http.StatusInternalServerError)
			log.Println("Error executing template:", err)
		}
	})

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Secrets represents the structure of the secrets.yaml file.
type Secrets struct {
	APIKeys map[string]string `yaml:"api_keys"`
}

// loadApiKey loads the API key for a given domain from a YAML file
func LoadApiKey(filePath, domain string) (string, error) {
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
