package timeutils

import (
	"time"
)



// determineRangeBasedOnCurrentDay calculates the range of days to consider based on the current day
func DetermineRangeBasedOnCurrentDay(currentDay time.Weekday) (time.Weekday, time.Weekday) {

//Final day data isn't available till 12pm on most days
now := time.Now()
onePM := time.Date(now.Year(), now.Month(), now.Day(), 13, 0, 0, 0, now.Location())

  switch currentDay {
	case time.Sunday:
    if now.Before(onePM){
    return time.Monday, time.Thursday
    } else{    
		return time.Monday, time.Friday
    }
	case time.Monday:
    if now.Before(onePM){
    return time.Tuesday, time.Friday
    } else{    
		return time.Tuesday, time.Saturday
    }
	case time.Tuesday:
    if now.Before(onePM){
    return time.Wednesday, time.Saturday
    } else{    
		return time.Wednesday, time.Sunday
    }
	case time.Wednesday:
    if now.Before(onePM){
    return time.Thursday, time.Sunday
    } else{    
		return time.Thursday, time.Monday
    }
	case time.Thursday:
    if now.Before(onePM){
    return time.Friday, time.Monday
    } else{    
		return time.Friday, time.Tuesday
    }
	case time.Friday:
    if now.Before(onePM){
    return time.Saturday, time.Tuesday
    } else{    
		return time.Saturday, time.Wednesday
    }
	case time.Saturday:
    if now.Before(onePM){
    return time.Sunday, time.Wednesday
    } else{    
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
    } else   if currentDay == time.Sunday {
        daysOrder = []time.Weekday{ time.Sunday, time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday}
      } else {
        daysOrder = []time.Weekday{time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday, time.Monday, time.Tuesday}
    }

    return daysOrder, startDay, endDay
}