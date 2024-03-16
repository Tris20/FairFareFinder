package weather_pleasantry

import (
	"github.com/Tris20/FairFareFinder/src/go_files"
	"time"
)
// determineRangeBasedOnCurrentDay, filterDataByDayRange, shouldIncludeDay defined here.

// determineRangeBasedOnCurrentDay calculates the range of days to consider based on the current day
func DetermineRangeBasedOnCurrentDay(currentDay time.Weekday) (time.Weekday, time.Weekday) {
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
func filterDataByDayRange(weeklyData []model.WeatherData, startDay, endDay time.Weekday) map[time.Weekday][]model.WeatherData {
	dailyData := make(map[time.Weekday][]model.WeatherData)
	for _, data := range weeklyData {
		timestamp := time.Unix(data.Dt, 0)
		day := timestamp.Weekday()
		hour := timestamp.Hour()

		if ShouldIncludeDay(day, startDay, endDay) {
			if hour >= 9 && hour <= 21 { // Include data points between 9 am and 9 pm
				dailyData[day] = append(dailyData[day], data)
			}
		}
	}
	return dailyData
}

// shouldIncludeDay checks if a day is within the specified range
func ShouldIncludeDay(day, startDay, endDay time.Weekday) bool {
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
