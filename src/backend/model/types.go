
package model

import "time"

type WeatherData struct {
	Dt   int64 `json:"dt"` // Unix timestamp of the forecasted data
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
}

// AirportInfo holds the details for an airport.
type DestinationInfo struct {
	IATA                string
	City                string
	Country             string
	SkyScannerURL       string
	SkyScannerID        string
	SkyScannerPrice     float64
	SkyScannerNextPrice float64
	AirbnbURL           string
	BookingURL          string
	WPI                 float64
	WeatherDetails      []DailyWeatherDetails
}

// AirportInfo holds the details for an airport.
type OriginInfo struct {
	IATA                   string
	City                   string
	Country                string
	DepartureStartDate     string
	DepartureEndDate       string
	ArrivalStartDate       string
	ArrivalEndDate         string
	NextDepartureStartDate string
	NextDepartureEndDate   string
	NextArrivalStartDate   string
	NextArrivalEndDate     string
	SkyScannerID           string
}

type DailyWeatherDetails struct {
	AverageTemp   float64
	CommonWeather string
	WPI           float64
	AverageWind   float64
	Icon          string
	Day           time.Weekday
}

// PriceData now just holds the price, as the IDs are embedded in the key of the map.
type PriceData struct {
	Price float64
}

type PersistentPrices struct {
	Data map[PriceKey]PriceData
}

type PriceKey struct {
	OriginID      string
	DestinationID string
}


type WeatherPleasantnessConfig struct {
	Conditions map[string]float64 `yaml:"conditions"`
}

// Location struct to hold unique location data
type Location struct {
	CityName      string
	CountryCode   string
	IATA          string
	SkyScannerID  string
	AirbnbURL     string
	BookingURL    string
	ThingsToDo    string
	FiveDayWPI    float64
}


// WeatherRecord holds weather data
type WeatherRecord struct {
	WeatherID        int
	CityName         string
	CountryCode      string
  IATA             string
	Date             string
	WeatherType      string
	Temperature      float64
	WeatherIconURL   string
	GoogleWeatherLink string
	WindSpeed        float64
  WPI              float64
}


// FlightPrice struct to hold flight price data
type FlightPrice struct {
    ID                    int
    Origin                string
    Destination           string
    OriginIATA            string
    DestinationIATA       string
    PriceThisWeek         float64
    SkyscannerURLThisWeek string
    PriceNextWeek         float64
    SkyscannerURLNextWeek string
}
