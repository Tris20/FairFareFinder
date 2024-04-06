package main

import (
	"fmt"
	"time"
)

// GetNextWeekday finds the next occurrence of a specific weekday from a given day.
// If `includeCurrent` is true and `baseDay` is the same as `weekday`, `baseDay` will be returned.
func GetNextWeekday(baseDay time.Time, weekday time.Weekday, includeCurrent bool) time.Time {
	daysAhead := int(weekday - baseDay.Weekday())
	if daysAhead < 0 || (!includeCurrent && daysAhead == 0) {
		daysAhead += 7
	}
	return baseDay.AddDate(0, 0, daysAhead)
}

// CalculateDateRange generates a string representation of the start and end dates,
// starting from `baseDay` and lasting for `duration` days.
func CalculateDateRange(baseDay time.Time, duration int) (startDate, endDate string) {
	startDate = baseDay.Format("2006-01-02")
	endDate = baseDay.AddDate(0, 0, duration-1).Format("2006-01-02")
	return
}

// CalculateWeekendRange calculates the date range for the specified weekend
// based on the current date and the `weekOffset` indicating which weekend to calculate.
// 0 = this weekend, 1 = next weekend, etc.
func CalculateWeekendRange(weekOffset int) (departureStartDate, departureEndDate, arrivalStartDate, arrivalEndDate string) {
	now := time.Now()
	// Adjust the base day according to the weekOffset
	baseDay := now.AddDate(0, 0, 7*weekOffset)
	
	// Find the upcoming Wednesday from baseDay
	upcomingWednesday := GetNextWeekday(baseDay, time.Wednesday, weekOffset == 0)
	// Calculate Wednesday to Saturday for that week
	departureStartDate, departureEndDate = CalculateDateRange(upcomingWednesday, 4)

	// Calculate Sunday to Wednesday for the weekend after the upcoming Wednesday
	nextSunday := upcomingWednesday.AddDate(0, 0, 4) // The Sunday after the upcoming Wednesday
	arrivalStartDate, arrivalEndDate = CalculateDateRange(nextSunday, 4)

	return
}

func main() {
	// Example usage:
	departureStartDate, departureEndDate, arrivalStartDate, arrivalEndDate := CalculateWeekendRange(0)
	fmt.Printf("This Weekend Departure: %s to %s\n", departureStartDate, departureEndDate)
	fmt.Printf("This Weekend Arrival: %s to %s\n", arrivalStartDate, arrivalEndDate)

	departureStartDate, departureEndDate, arrivalStartDate, arrivalEndDate = CalculateWeekendRange(1)
	fmt.Printf("Next Weekend Departure: %s to %s\n", departureStartDate, departureEndDate)
	fmt.Printf("Next Weekend Arrival: %s to %s\n", arrivalStartDate, arrivalEndDate)

	departureStartDate, departureEndDate, arrivalStartDate, arrivalEndDate = CalculateWeekendRange(2)
	fmt.Printf("Weekend After Next Departure: %s to %s\n", departureStartDate, departureEndDate)
	fmt.Printf("Weekend After Next Arrival: %s to %s\n", arrivalStartDate, arrivalEndDate)


}

