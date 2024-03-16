package timeutils

import (
	"time"
)



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
    } else {
        daysOrder = []time.Weekday{time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday, time.Monday, time.Tuesday}
    }

    return daysOrder, startDay, endDay
}
