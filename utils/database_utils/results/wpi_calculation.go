
package main

import (
	"FairFareFinder/src/go_files/timeutils"
	"strings"
	"time"
	"log"
    "fmt"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

func ProcessLocation(location Location, records []WeatherRecord) (float64, map[time.Weekday]WeatherRecord) {
	// Fetch weather data from source database
/*
	// Iterate over weatherData and print each element (for debugging)
	for _, data := range weatherData {
		fmt.Printf("CityName: %s, CountryCode: %s, Date: %s, WeatherType: %s, Temperature: %.1f, WindSpeed: %.1f, WPI: %.1f, WeatherIconURL: %s, GoogleWeatherLink: %s\n",
			data.CityName, data.CountryCode, data.Date, data.WeatherType, data.Temperature, data.WindSpeed, data.WPI, data.WeatherIconURL, data.GoogleWeatherLink)
	}
*/

  weatherData := FilterWeatherRecords(location, records)
	// Calculate daily average, and next 5 days average WPI
	dailyAverageWeather, next5DaysAverageWPI := CalculateWPI(weatherData)

	// Debug prints
	fmt.Printf("Location: %s, %s (%s)\n", location.CityName, location.CountryCode, location.IATA)
	fmt.Printf("Next 5 Days Average WPI: %.2f\n", next5DaysAverageWPI)
	fmt.Println("Daily Average Weather Details:")
	for day, details := range dailyAverageWeather {
		fmt.Printf("Day: %s, Temp: %.2f, Weather: %s, Wind: %.2f, WPI: %.2f\n",
			day, details.Temperature, details.WeatherType, details.WindSpeed, details.WPI)
	}

	return next5DaysAverageWPI, dailyAverageWeather


}

func CalculateWPI(weeklyData []WeatherRecord) (map[time.Weekday]WeatherRecord, float64) {
	config, err := LoadWeatherPleasantnessConfig("config/weatherPleasantness.yaml")
	if err != nil {
		log.Fatal("Error loading weather pleasantness config:", err)
	}

	currentDay := time.Now().Weekday()
	startDay, endDay := timeutils.DetermineRangeBasedOnCurrentDay(currentDay)

	dailyData := FilterDataByDayRange(weeklyData, startDay, endDay)
	dailyDetails := make(map[time.Weekday]WeatherRecord)
	var totalWPI float64

	for day, data := range dailyData {
		var sumTemp, sumWind, count float64
		weatherCount := make(map[string]int)
		var maxWeather string
		var maxCount int

		var icon string
		for _, segment := range data {
			sumTemp += segment.Temperature
			sumWind += segment.WindSpeed
			weatherCount[segment.WeatherType]++
			if weatherCount[segment.WeatherType] > maxCount {
				maxCount = weatherCount[segment.WeatherType]
				maxWeather = segment.WeatherType
			}
			count++
		}

		if count == 0 {
			fmt.Printf("No data for day: %s\n", day)
			continue
		}

		if len(data) >= 2 {
			icon = data[1].WeatherIconURL
		} else {
			icon = data[0].WeatherIconURL
		}
		icon = strings.Replace(icon, "n", "d", 1)

		avgWind := sumWind / count
		avgTemp := sumTemp / count
		wpi := calculateDailyAverageWPI(data, config)

		// Debug prints for daily calculations
		fmt.Printf("Day: %s, Count: %.2f, SumTemp: %.2f, SumWind: %.2f, AvgTemp: %.2f, AvgWind: %.2f, WPI: %.2f\n",
			day, count, sumTemp, sumWind, avgTemp, avgWind, wpi)


		dailyDetails[day] = WeatherRecord{
			CityName:         data[0].CityName,
			CountryCode:      data[0].CountryCode,
			IATA:             data[0].IATA,
			Date:             data[0].Date,
			WeatherType:      maxWeather,
			Temperature:      avgTemp,
			WeatherIconURL:   icon,
			GoogleWeatherLink: data[0].GoogleWeatherLink,
			WindSpeed:        avgWind,
			WPI:              wpi,
		}
		totalWPI += wpi
	}

	if len(dailyDetails) == 0 {
		fmt.Println("No daily details calculated.")
		return dailyDetails, 0
	}

	averageWPI := totalWPI / float64(len(dailyDetails))

fmt.Printf("TotalWPI: %.2f, Len(dailyDetails): %d, AverageWPI: %.2f\n", totalWPI, len(dailyDetails), averageWPI)
  return dailyDetails, averageWPI
}

func calculateDailyAverageWPI(weatherData []WeatherRecord, config WeatherPleasantnessConfig) float64 {
	var totalWPI float64
	var count float64

	for _, data := range weatherData {
		wpi := weatherPleasantness(data.Temperature, data.WindSpeed, data.WeatherType, config)
		totalWPI += wpi
		count++
	}

	if count == 0 {
		return 0
	}

	return totalWPI / count
}

func weatherPleasantness(temp float64, wind float64, cond string, config WeatherPleasantnessConfig) float64 {
	weightTemp := 5.0
	weightWind := 1.0
	weightCond := 2.0

	tempIndex := tempPleasantness(temp) * weightTemp
	windIndex := windPleasantness(wind) * weightWind
	weatherIndex := weatherCondPleasantness(cond, config) * weightCond

	index := (tempIndex + windIndex + weatherIndex) / (weightTemp + weightWind + weightCond)
	return index
}

func LoadWeatherPleasantnessConfig(filePath string) (WeatherPleasantnessConfig, error) {
	var config WeatherPleasantnessConfig
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(yamlFile, &config)
	return config, err
}

