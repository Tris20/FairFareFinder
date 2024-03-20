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
	} else {
		daysOrder = []time.Weekday{time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday, time.Monday, time.Tuesday}
	}

	return daysOrder, startDay, endDay
}

// ListDatesBetween generates a list of dates in the format "YYYY-MM-DD" between two given dates, inclusive.
func ListDatesBetween(start string, end string) ([]string, error) {
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
		fmt.Printf("\n %s", d.Format(layout))
		dates = append(dates, d.Format(layout))
	}

	return dates, nil
}





// formatDate formats a time.Time object into a string in the "YYYY-MM-DD" format.
func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// upcomingWedToSat calculates and returns the date range for the upcoming Wednesday to Saturday.
func UpcomingWedToSat() (string, string) {
	return calculateDateRange(time.Wednesday, 4)
}

// upcomingSunToWed calculates and returns the date range for the upcoming Sunday to Wednesday.
func UpcomingSunToWed() (string, string) {
	return calculateDateRange(time.Sunday, 4)
}

// calculateDateRange calculates the date range starting from the next occurrence of startDayOfWeek and lasting for a specified number of days.
// It returns the start date and end date of the range.
func calculateDateRange(startDayOfWeek time.Weekday, durationDays int) (string, string) {
	now := time.Now()
	// Calculate the number of days until the next occurrence of the start day of the week
	daysUntilStart := (int(startDayOfWeek) - int(now.Weekday()) + 7) % 7
	if daysUntilStart == 0 {
		daysUntilStart = 7 // If today is the start day, begin from the next occurrence
	}
	// Calculate the start date by adding daysUntilStart to the current date
	startDate := now.AddDate(0, 0, daysUntilStart)
	// Calculate the end date by adding durationDays - 1 to the start date
	endDate := startDate.AddDate(0, 0, durationDays-1)

	return formatDate(startDate), formatDate(endDate)
}
