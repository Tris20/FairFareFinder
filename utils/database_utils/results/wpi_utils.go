package main 

// Simple math functions and checks go here

func interpolate(temp, temp1, temp2, score1, score2 float64) float64 {
	return ((temp-temp1)/(temp2-temp1))*(score2-score1) + score1
}

// tempPleasantness, windPleasantness, weatherCondPleasantness defined here.

func tempPleasantness(temperature float64) float64 {
	// Optimal range
	if temperature >= 22 && temperature <= 26 {
		return 10
	}

	//     Interpolation between key temperatures below the optimal range
	if temperature > 18 && temperature < 22 {
		return interpolate(temperature, 18, 22, 7, 10)
	}

	// Interpolation between key temperatures above the optimal range
	if temperature > 26 && temperature < 40 {
		return interpolate(temperature, 26, 40, 10, 0)
	}

	// Below 18 down to 0
	if temperature >= 5 && temperature <= 18 {
		return interpolate(temperature, 5, 18, 0, 7)
	}

	// Anything below 0 or above 40
	if temperature <= 5 || temperature >= 40 {
		return 0
	}

	return 0 // Default case if needed
}

// windPleasantness returns a value between 0 and 10 for wind condition pleasantness
func windPleasantness(windSpeed float64) float64 {
	worstWind := 13.8
	if windSpeed >= worstWind {
		return 0
	} else {
		return 10 - windSpeed*10/worstWind
	}
}

// weatherCondPleasantness returns a value between 0 and 10 for weather condition pleasantness
func weatherCondPleasantness(cond string, config WeatherPleasantnessConfig) float64 {
	pleasantness, ok := config.Conditions[cond]
	if !ok {
		return 0
	}
	return pleasantness
}


// filterDataByDayRange filters the weather data for a specific range of days
func filterDataByDayRange(weeklyData []model.WeatherData, startDay, endDay time.Weekday) map[time.Weekday][]model.WeatherData {
	dailyData := make(map[time.Weekday][]model.WeatherData)
	for _, data := range weeklyData {
		timestamp := time.Unix(data.Dt, 0)
		day := timestamp.Weekday()
		hour := timestamp.Hour()

		if timeutils.ShouldIncludeDay(day, startDay, endDay) {
			if hour >= 9 && hour <= 21 { // Include data points between 9 am and 9 pm
				dailyData[day] = append(dailyData[day], data)
			}
		}
	}
	return dailyData
}
