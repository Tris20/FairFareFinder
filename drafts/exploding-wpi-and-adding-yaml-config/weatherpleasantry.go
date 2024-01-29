
package main

import (
	"io/ioutil"
	"time"
	"gopkg.in/yaml.v2"
)

// WeatherPleasantnessConfig holds the configuration for weather pleasantness ratings.
type WeatherPleasantnessConfig struct {
	Conditions map[string]float64 `yaml:"conditions"`
}

type DailyWeatherDetails struct {
    AverageTemp  float64
    CommonWeather string
    WPI          float64
}

// LoadWeatherPleasantnessConfig loads the weather pleasantness configuration from a YAML file.
func LoadWeatherPleasantnessConfig(filePath string) (WeatherPleasantnessConfig, error) {
	var config WeatherPleasantnessConfig

	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

// weatherPleasantness calculates the "weather pleasentness index" (WPI)
func weatherPleasantness(temp float64, wind float64, cond string, config WeatherPleasantnessConfig) float64 {
	weightTemp := 3.0
	weightWind := 1.0
	weightCond := 2.0

	tempIndex := tempPleasantness(temp) * weightTemp
	windIndex := windPleasantness(wind) * weightWind
	weatherIndex := weatherCondPleasantness(cond, config) * weightCond

	index := (tempIndex + windIndex + weatherIndex) / (weightTemp + weightWind + weightCond)
	return index
}

// tempPleasantness returns a value between 0 and 10 for temperature pleasantness
func tempPleasantness(temperature float64) float64 {
	GoodTemp := 20.0
	indexAtGoodTemp := 7.0
	PerfectTemp := 23.0
	slope := indexAtGoodTemp / GoodTemp

	if temperature <= 0 {
		return 0
	} else if temperature > PerfectTemp {
		return 10
	} else {
		return slope * temperature
	}
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

// calculateDailyAverageWPI calculates the average WPI for a single day
// This function assumes it receives weather data for each 3-hour segment between 9 am and 9 pm
func calculateDailyAverageWPI(weatherData []WeatherData, config WeatherPleasantnessConfig) float64 {
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


//func ProcessForecastData(weeklyData []WeatherData, config WeatherPleasantnessConfig) (map[time.Weekday]DailyWeatherDetails, float64){
    
func ProcessForecastData(weeklyData []WeatherData, config WeatherPleasantnessConfig) (map[time.Weekday]DailyWeatherDetails, float64) {
    currentDay := time.Now().Weekday()
    startDay, endDay := determineRangeBasedOnCurrentDay(currentDay)

    dailyData := filterDataByDayRange(weeklyData, startDay, endDay)
//  dailyData := make(map[time.Weekday][]WeatherData)
//    for _, data := range weeklyData {
//        timestamp := time.Unix(data.Dt, 0)
//        day := timestamp.Weekday()
//        hour := timestamp.Hour()
//
//    //  fmt.Printf("Day info %s, Hour: %d\n", day.String(), hour)
//
//        if day >= time.Thursday && day <= time.Saturday {
//            // Only include data points between 9 am and 9 pm
//            if hour >= 9 && hour <= 21 {
//                dailyData[day] = append(dailyData[day], data)
//            }
//        }
//    }
//
    dailyDetails := make(map[time.Weekday]DailyWeatherDetails)
    var totalWPI float64
    for day, data := range dailyData {
        var sumTemp, count float64
        weatherCount := make(map[string]int)
        var maxWeather string
        var maxCount int

        for _, segment := range data {
            sumTemp += segment.Main.Temp
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

        avgTemp := sumTemp / count
        wpi := calculateDailyAverageWPI(data, config)
        dailyDetails[day] = DailyWeatherDetails{
            AverageTemp:  avgTemp,
            CommonWeather: maxWeather,
            WPI:          wpi,
        }
        totalWPI += wpi
    }

    averageWPI := totalWPI / float64(len(dailyDetails))
    return dailyDetails, averageWPI
}


// determineRangeBasedOnCurrentDay calculates the range of days to consider based on the current day
func determineRangeBasedOnCurrentDay(currentDay time.Weekday) (time.Weekday, time.Weekday) {
    switch currentDay {
    case time.Sunday:
        return time.Wednesday, time.Friday
    case time.Monday:
        return time.Wednesday, time.Saturday
    case time.Tuesday:
        return time.Thursday, time.Sunday
    case time.Wednesday:
        return time.Thursday, time.Monday
    case time.Thursday:
        return time.Friday, time.Tuesday
    case time.Friday:
        return time.Saturday, time.Wednesday
    case time.Saturday:
        return time.Sunday, time.Thursday
    default:
        return time.Thursday, time.Monday // Default range
    }
}

// filterDataByDayRange filters the weather data for a specific range of days
func filterDataByDayRange(weeklyData []WeatherData, startDay, endDay time.Weekday) map[time.Weekday][]WeatherData {
    dailyData := make(map[time.Weekday][]WeatherData)
    for _, data := range weeklyData {
        timestamp := time.Unix(data.Dt, 0)
        day := timestamp.Weekday()
        hour := timestamp.Hour()

        if shouldIncludeDay(day, startDay, endDay) {
            if hour >= 9 && hour <= 21 { // Include data points between 9 am and 9 pm
                dailyData[day] = append(dailyData[day], data)
            }
        }
    }
    return dailyData
}

// shouldIncludeDay checks if a day is within the specified range
func shouldIncludeDay(day, startDay, endDay time.Weekday) bool {
    for d := startDay; d != endDay; d = (d + 1) % 7 {
        if d == day {
            return true
        }
        if d == endDay {
            break
        }
    }
    return day == endDay
}
