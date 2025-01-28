package backend

import "database/sql"

// Keep basic structs in a models file

type Weather struct {
	Date           string
	AvgDaytimeTemp sql.NullFloat64
	WeatherIcon    string
	GoogleUrl      string
	AvgDaytimeWpi  sql.NullFloat64
}

type Flight struct {
	DestinationCityName  string
	RandomImageURL       string
	PriceCity1           sql.NullFloat64
	UrlCity1             string
	WeatherForecast      []Weather
	AvgWpi               sql.NullFloat64
	BookingUrl           sql.NullString
	BookingPppn          sql.NullFloat64
	FiveNightsFlights    sql.NullFloat64
	DurationMins         sql.NullInt64
	DurationHours        sql.NullInt64
	DurationHoursRounded sql.NullInt64
	DurationHourDotMins  sql.NullFloat64
}

type FlightsData struct {
	SelectedCity1          string
	Flights                []Flight
	MaxWpi                 sql.NullFloat64
	MinFlight              sql.NullFloat64
	MinHotel               sql.NullFloat64
	MinFnaf                sql.NullFloat64
	AllAccommodationPrices []float64
}
