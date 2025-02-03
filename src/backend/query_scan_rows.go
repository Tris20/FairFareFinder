package backend

import (
	"database/sql"
	"github.com/Tris20/FairFareFinder/src/backend/model"
	"log"
)

// buildFlightsData builds the data for the template from cities, flights, and accommodations.
func buildFlightsData(cities []string, flights []model.Flight) model.FlightsData {
	// Ensure there is at least one city in the list
	var selectedCity1 string
	if len(cities) > 0 {
		selectedCity1 = cities[0]
	} else {
		selectedCity1 = "" // Default to an empty string if no cities are provided
	}

	// Initialize variables for max/min values
	var maxWpi, minFlightPrice, minHotelPrice, minFnafPrice sql.NullFloat64

	// Process each flight to find max/min values
	for _, flight := range flights {
		maxWpi = UpdateMaxValue(maxWpi, flight.AvgWpi)
		minFlightPrice = UpdateMinValue(minFlightPrice, flight.PriceCity1)
		minHotelPrice = UpdateMinValue(minHotelPrice, flight.BookingPppn)
		minFnafPrice = UpdateMinValue(minFnafPrice, flight.FiveNightsFlights)

	}

	// Build and return the FlightsData
	return model.FlightsData{
		SelectedCity1: selectedCity1,
		Flights:       flights,
		MaxWpi:        maxWpi,
		MinFlight:     minFlightPrice,
		MinHotel:      minHotelPrice,
		MinFnaf:       minFnafPrice,
	}
}

func BuildTemplateData(cities []string, flights []model.Flight, allAccomPrices []float64, allFlightPrices []float64) model.FlightsData {
	data := buildFlightsData(cities, flights)
	data.AllAccommodationPrices = allAccomPrices
	data.AllFlightPrices = allFlightPrices
	return data
}

// Helper function to process rows into flight and weather data
func ProcessFlightRows(rows *sql.Rows) ([]model.Flight, error) {
	var flights []model.Flight
	for rows.Next() {
		var flight model.Flight
		var weather model.Weather
		var imageUrl sql.NullString
		var bookingUrl sql.NullString
		var priceFnaf sql.NullFloat64
		var duration_mins sql.NullInt64
		var duration_hours sql.NullInt64
		var duration_hours_rounded sql.NullInt64
		var duration_hour_dot_mins sql.NullFloat64

		err := rows.Scan(
			&flight.DestinationCityName,
			&flight.PriceCity1,
			&flight.UrlCity1,
			&weather.Date,
			&weather.AvgDaytimeTemp,
			&weather.WeatherIcon,
			&weather.GoogleUrl,
			&flight.AvgWpi,
			&imageUrl,
			&bookingUrl,
			&flight.BookingPppn,
			&priceFnaf,
			&duration_mins,
			&duration_hours,
			&duration_hours_rounded,
			&duration_hour_dot_mins,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}

		SetFlightDurationInt(&flight, duration_mins, &flight.DurationMins, "Duration: %d minutes for flight to %s")
		SetFlightDurationInt(&flight, duration_hours, &flight.DurationHours, "Duration: %d hours for flight to %s")
		SetFlightDurationInt(&flight, duration_hours_rounded, &flight.DurationHoursRounded, "Duration: %d rounded hours for flight to %s")
		SetFlightDurationFloat(&flight, duration_hour_dot_mins, &flight.DurationHourDotMins, "Duration: %.2f hours.mins for flight to %s")

		// Log the weather data for debugging
		log.Printf("Row Data - Destination: %s, Date: %s, Temp: %.2f, Icon: %s, Duration.Hours: %d, Duration.Mins: %d ",
			flight.DestinationCityName,
			weather.Date,
			weather.AvgDaytimeTemp.Float64,
			weather.WeatherIcon,
			flight.DurationHours.Int64,
			flight.DurationMins.Int64,
		)

		// Log the imageUrl for debugging
		log.Printf("Scanned image URL: '%s', Valid: %t", imageUrl.String, imageUrl.Valid)

		if imageUrl.Valid && len(imageUrl.String) > 5 {
			flight.RandomImageURL = imageUrl.String
			log.Printf("Using image URL from database: %s", flight.RandomImageURL)
		} else {
			flight.RandomImageURL = "/images/location-placeholder-image.png"
			log.Printf("Using default placeholder image URL: %s", flight.RandomImageURL)
		}
		flight.BookingUrl = bookingUrl
		flight.FiveNightsFlights = priceFnaf
		addOrUpdateFlight(&flights, flight, weather)
	}
	return flights, nil
}

// Helper function to add or update flight entries
func addOrUpdateFlight(flights *[]model.Flight, flight model.Flight, weather model.Weather) {
	for i := range *flights {
		if (*flights)[i].DestinationCityName == flight.DestinationCityName {
			(*flights)[i].WeatherForecast = append((*flights)[i].WeatherForecast, weather)
			return
		}
	}

	flight.WeatherForecast = []model.Weather{weather}
	*flights = append(*flights, flight)
}
