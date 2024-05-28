
// wpi_utils.go

package main

import (
	"time"
  "fmt"
	"FairFareFinder/src/go_files/timeutils"
)

// Simple math functions and checks go here

func interpolate(temp, temp1, temp2, score1, score2 float64) float64 {
	return ((temp-temp1)/(temp2-temp1))*(score2-score1) + score1
}

func tempPleasantness(temperature float64) float64 {
	if temperature >= 22 && temperature <= 26 {
		return 10
	}
	if temperature > 18 && temperature < 22 {
		return interpolate(temperature, 18, 22, 7, 10)
	}
	if temperature > 26 && temperature < 40 {
		return interpolate(temperature, 26, 40, 10, 0)
	}
	if temperature >= 5 && temperature <= 18 {
		return interpolate(temperature, 5, 18, 0, 7)
	}
	if temperature <= 5 || temperature >= 40 {
		return 0
	}
	return 0
}

func windPleasantness(windSpeed float64) float64 {
	worstWind := 13.8
	if windSpeed >= worstWind {
		return 0
	} else {
		return 10 - windSpeed*10/worstWind
	}
}

func weatherCondPleasantness(cond string, config WeatherPleasantnessConfig) float64 {
	pleasantness, ok := config.Conditions[cond]
	if !ok {
		return 0
	}
	return pleasantness
}



func FilterDataByDayRange(weeklyData []WeatherRecord, startDay, endDay time.Weekday) map[time.Weekday][]WeatherRecord {
	dailyData := make(map[time.Weekday][]WeatherRecord)
	for _, data := range weeklyData {
		timestamp, err := time.Parse("2006-01-02 15:04:05", data.Date)
		if err != nil {
			fmt.Printf("Error parsing date: %v\n", err)
			continue
		}
		day := timestamp.Weekday()
		hour := timestamp.Hour()

		if timeutils.ShouldIncludeDay(day, startDay, endDay) {
			if hour >= 9 && hour <= 21 {
				dailyData[day] = append(dailyData[day], data)
			}
		}
	}
	fmt.Printf("Filtered Data By Day Range: %v\n", dailyData)
	return dailyData
}



func FilterWeatherRecords(location Location, records []WeatherRecord) []WeatherRecord {
	var filteredRecords []WeatherRecord
	for _, record := range records {
		if record.CityName == location.CityName && record.CountryCode == location.CountryCode && record.IATA == location.IATA {
			filteredRecords = append(filteredRecords, record)
		}
	}
	fmt.Printf("Filtered Weather Records for %s, %s (%s): %v\n", location.CityName, location.CountryCode, location.IATA, filteredRecords)
	return filteredRecords
}



func CollectDailyAverageWeather(records []WeatherRecord) ([]WeatherRecord, error) {
	// Extract unique locations
	uniqueLocations := getUniqueLocations(records)

	var dailyAverageWeatherRecords []WeatherRecord

	for _, loc := range uniqueLocations {
		// Process location to get daily average weather
		_, dailyAverageWeather := ProcessLocation(loc, records)

		// Collect all daily average weather records
		for _, weatherRecord := range dailyAverageWeather {
			dailyAverageWeatherRecords = append(dailyAverageWeatherRecords, weatherRecord)
		}
	}

	return dailyAverageWeatherRecords, nil
}
