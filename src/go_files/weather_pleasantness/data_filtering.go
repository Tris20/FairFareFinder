package weather_pleasantry

import (
	"github.com/Tris20/FairFareFinder/src/backend"
	"github.com/Tris20/FairFareFinder/src/go_files/timeutils"
	"time"
)

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
