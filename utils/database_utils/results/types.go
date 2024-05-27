package main

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


