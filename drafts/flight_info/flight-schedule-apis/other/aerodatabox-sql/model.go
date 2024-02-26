
package main

// APIResponse represents the structure of the API response for departures
type APIResponse struct {
	Departures []Departure `json:"departures"`
}

// Departure contains the detailed information of a flight's departure and arrival
type Departure struct {
	Departure Detail `json:"departure"`
	Arrival   Detail `json:"arrival"`
}

// Detail holds the information about either departure or arrival details of a flight
type Detail struct {
	Airport       Airport `json:"airport"`
	ScheduledTime Time    `json:"scheduledTime"`
}

// Airport provides the IATA code of the airport
type Airport struct {
	IATA string `json:"iata"`
}

// Time represents the scheduled times for departure or arrival
type Time struct {
	Local string `json:"local"`
}
