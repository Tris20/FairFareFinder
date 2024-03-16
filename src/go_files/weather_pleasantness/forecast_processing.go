package weather_pleasantry

import (
	"github.com/Tris20/FairFareFinder/src/go_files/timeutils"
	"github.com/Tris20/FairFareFinder/src/go_files"
	"strings"
	"time"
)


// calculateDailyAverageWPI and ProcessForecastData defined here.
// calculateDailyAverageWPI calculates the average WPI for a single day
// This function assumes it receives weather data for each 3-hour segment between 9 am and 9 pm
func calculateDailyAverageWPI(weatherData []model.WeatherData, config WeatherPleasantnessConfig) float64 {
	var totalWPI float64
	var count float64

	for _, data := range weatherData {
		// Assuming WeatherData contains Temp, Wind.Speed, and Weather[0].Main
		wpi := weatherPleasantness(data.Main.Temp, data.Wind.Speed, data.Weather[0].Main, config)
		totalWPI += wpi
		count++
	}

	if count == 0 {
		return 0
	}

	return totalWPI / count
}




// ProcessForecastData takes a slice of WeatherData for an entire week
// and returns a map of average WPI for Thursday to Monday.
// It also calculates the overall average for these days.
// Assuming each WeatherData entry is for a 3-hour segment

func ProcessForecastData(weeklyData []model.WeatherData, config WeatherPleasantnessConfig) (map[time.Weekday]model.DailyWeatherDetails, float64) {

  currentDay := time.Now().Weekday()
	startDay, endDay := timeutils.DetermineRangeBasedOnCurrentDay(currentDay)

	dailyData := filterDataByDayRange(weeklyData, startDay, endDay)
	dailyDetails := make(map[time.Weekday]model.DailyWeatherDetails)
	var totalWPI float64
	// Assuming this part needs correction
	for day, data := range dailyData {
		var sumTemp, sumWind, count float64
		weatherCount := make(map[string]int)
		var maxWeather string
		var maxCount int

		var icon string
		for _, segment := range data {
			sumTemp += segment.Main.Temp
			sumWind += segment.Wind.Speed // Correctly access Wind.Speed here
			weatherCount[segment.Weather[0].Main]++
			if weatherCount[segment.Weather[0].Main] > maxCount {
				maxCount = weatherCount[segment.Weather[0].Main]
				maxWeather = segment.Weather[0].Main
			}
			count++
		}

		if count == 0 {
			continue
		}

		//Get the condition code from roughly mid day
		if len(data) >= 2 {
			// Access the second segment directly
			icon = data[1].Weather[0].Icon
		} else {
			icon = data[0].Weather[0].Icon
		}
		icon = strings.Replace(icon, "n", "d", 1) //replace night icons with day equivalent

		avgWind := sumWind / count // Calculate average wind here
		avgTemp := sumTemp / count
		wpi := calculateDailyAverageWPI(data, config)

		// Create the weather entry for that day in dailyDetails
		dailyDetails[day] = model.DailyWeatherDetails{
			AverageTemp:   avgTemp,
			CommonWeather: maxWeather,
			WPI:           wpi,
			AverageWind:   avgWind, // Use the calculated avgWind
			Icon:          icon,
			Day:           day,
		}
		totalWPI += wpi
	}
	averageWPI := totalWPI / float64(len(dailyDetails))
	return dailyDetails, averageWPI
}




// weatherPleasantness calculates the "weather pleasentness index" (WPI)
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

