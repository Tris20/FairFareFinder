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

func UpcomingWedToSat() (string, string) {
	return calculateDateRangeForWedToSat(time.Wednesday, 4)
}

func UpcomingSunToWedFromSat(endSaturday string) (string, string) {
	// Parse the Saturday date
	satDate, _ := time.Parse("2006-01-02", endSaturday)
	// Calculate Sunday to Wednesday from that Saturday
	return calculateDateRangeForSunToWed(satDate, 4)
}

// Adjusted for the requirement
func calculateDateRangeForWedToSat(startDayOfWeek time.Weekday, durationDays int) (string, string) {
	now := time.Now()
	// If today is between Wednesday and Saturday, include this week
	if now.Weekday() >= time.Wednesday && now.Weekday() <= time.Saturday {
		daysUntilStart := int(startDayOfWeek - now.Weekday())
		if daysUntilStart > 0 {
			daysUntilStart -= 7 // Move back to the current week's Wednesday
		}
		startDate := now.AddDate(0, 0, daysUntilStart)
		endDate := startDate.AddDate(0, 0, durationDays-1)
		return formatDate(startDate), formatDate(endDate)
	}
	// Else, find the next Wednesday to Saturday
	return calculateDateRange(startDayOfWeek, durationDays)
}

// Calculate Sunday to Wednesday after a given Saturday
func calculateDateRangeForSunToWed(afterDate time.Time, durationDays int) (string, string) {
	startDate := afterDate.AddDate(0, 0, 1) // Next day after Saturday, which is Sunday
	endDate := startDate.AddDate(0, 0, durationDays-1)
	return formatDate(startDate), formatDate(endDate)
}

// Generic calculateDateRange, assuming it might still be used elsewhere
func calculateDateRange(startDayOfWeek time.Weekday, durationDays int) (string, string) {
	now := time.Now()
	daysUntilStart := (int(startDayOfWeek) - int(now.Weekday()) + 7) % 7
	if daysUntilStart == 0 {
		daysUntilStart = 7
	}
	startDate := now.AddDate(0, 0, daysUntilStart)
	endDate := startDate.AddDate(0, 0, durationDays-1)
	return formatDate(startDate), formatDate(endDate)
}

// Helper function to format time.Time objects into strings
func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}
