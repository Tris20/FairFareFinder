package timeutils

import (
	"fmt"
	"time"
)

// determineRangeBasedOnCurrentDay calculates the range of days to consider based on the current day
func DetermineRangeBasedOnCurrentDay(currentDay time.Weekday) (time.Weekday, time.Weekday) {

	//Final day data isn't available till 12pm on most days
	now := time.Now()
	onePM := time.Date(now.Year(), now.Month(), now.Day(), 13, 0, 0, 0, now.Location())

	switch currentDay {
	case time.Sunday:
		if now.Before(onePM) {
			return time.Monday, time.Thursday
		} else {
			return time.Monday, time.Friday
		}
	case time.Monday:
		if now.Before(onePM) {
			return time.Tuesday, time.Friday
		} else {
			return time.Tuesday, time.Saturday
		}
	case time.Tuesday:
		if now.Before(onePM) {
			return time.Wednesday, time.Saturday
		} else {
			return time.Wednesday, time.Sunday
		}
	case time.Wednesday:
		if now.Before(onePM) {
			return time.Thursday, time.Sunday
		} else {
			return time.Thursday, time.Monday
		}
	case time.Thursday:
		if now.Before(onePM) {
			return time.Friday, time.Monday
		} else {
			return time.Friday, time.Tuesday
		}
	case time.Friday:
		if now.Before(onePM) {
			return time.Saturday, time.Tuesday
		} else {
			return time.Saturday, time.Wednesday
		}
	case time.Saturday:
		if now.Before(onePM) {
			return time.Sunday, time.Wednesday
		} else {
			return time.Sunday, time.Thursday
		}
	default:
		return time.Thursday, time.Monday // Default range
	}
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

// GetDaysOrder returns the order of days for the week based on the current day, along with startDay and endDay.
func GetDaysOrder() ([]time.Weekday, time.Weekday, time.Weekday) {
	var daysOrder []time.Weekday
	currentDay := time.Now().Weekday()

	startDay, endDay := DetermineRangeBasedOnCurrentDay(currentDay)

	if currentDay == time.Saturday {
		daysOrder = []time.Weekday{time.Saturday, time.Sunday, time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}
	} else if currentDay == time.Sunday {
		daysOrder = []time.Weekday{time.Sunday, time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday}
	} else if currentDay == time.Monday {
		daysOrder = []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday}
	} else if currentDay == time.Tuesday {
		daysOrder = []time.Weekday{time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday, time.Monday}
	} else if currentDay == time.Friday {
		daysOrder = []time.Weekday{time.Friday, time.Saturday, time.Sunday, time.Monday,
			time.Tuesday, time.Wednesday, time.Thursday}

	} else {
		daysOrder = []time.Weekday{time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday, time.Monday, time.Tuesday}
	}

	return daysOrder, startDay, endDay
}

// ListDatesBetween generates a list of dates in the format "YYYY-MM-DD" between two given dates, inclusive.
func ListDatesBetween(start, end string) ([]string, error) {
	const layout = "2006-01-02" // Go's reference time format
	startDate, err := time.Parse(layout, start)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %v", err)
	}

	endDate, err := time.Parse(layout, end)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %v", err)
	}

	var dates []string
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format(layout))
	}

	return dates, nil
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

// GetNextWeekday finds the next occurrence of a specific weekday from a given day.
// If `includeCurrent` is true and `baseDay` is the same as `weekday`, `baseDay` will be returned.
func GetNextWeekday(baseDay time.Time, weekday time.Weekday, includeCurrent bool) time.Time {
	daysAhead := int(weekday - baseDay.Weekday())
	if daysAhead < 0 || (!includeCurrent && daysAhead == 0) {
		daysAhead += 7
	}
	return baseDay.AddDate(0, 0, daysAhead)
}

// Helper function to format time.Time objects into strings.
func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}




