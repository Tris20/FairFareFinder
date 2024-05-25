
package main

import (
	"fmt"
	"time"
)

// Will use this later when doing flightprice table
/*
func GenerateSkyScannerURL(originIataCode string, destIataCode string) string {
	now := time.Now()
	formattedDate := now.Format("0601")
	baseSkyScannerURL := fmt.Sprintf("https://www.skyscanner.de/transport/fluge/%s/%s/?adults=1&adultsv2=1&cabinclass=economy&children=0&inboundaltsenabled=false&infants=0&outboundaltsenabled=false&preferdirects=true&ref=home&rtn=1&oym=%s", originIataCode, destIataCode, formattedDate)
	return baseSkyScannerURL
}
*/

// Function to generate the Airbnb URL with dynamic values
func GenerateAirbnbURL(cityName string) string {
	checkin := nextThursdayFridaySaturday()
	checkout := checkin.Add(3 * 24 * time.Hour) // Adding 3 days for a one-week stay
	return fmt.Sprintf("https://www.airbnb.de/s/%s/homes?adults=1&checkin=%s&checkout=%s&flexible_trip_lengths%%5B%%5D=one_week&price_filter_num_nights=3&price_max=112", cityName, checkin.Format("2006-01-02"), checkout.Format("2006-01-02"))
}

// Function to generate the Booking.com URL with dynamic values
func GenerateBookingURL(cityName string) string {
	checkin := nextThursdayFridaySaturday()
	checkout := checkin.Add(3 * 24 * time.Hour) // Adding 3 days for a one-week stay
	return fmt.Sprintf("https://www.booking.com/searchresults.en-gb.html?ss=%s&group_adults=1&no_rooms=1&group_children=0&nflt=price%%3DEUR-min-110-1%%3Breview_score%%3D80&flex_window=2&checkin=%s&checkout=%s", cityName, checkin.Format("2006-01-02"), checkout.Format("2006-01-02"))
}

// Function to get the date of the next Thursday, Friday, or Saturday from today's date
func nextThursdayFridaySaturday() time.Time {
	today := time.Now()
	for {
		switch today.Weekday() {
		case time.Thursday, time.Friday, time.Saturday:
			return today
		}
		today = today.Add(24 * time.Hour)
	}
}

