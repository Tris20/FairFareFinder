package main


func PrepareLocationData(records []WeatherRecord) ([]Location, error) {
	uniqueLocations := getUniqueLocations(records)

	// Collect unique IATA codes
	iataCodes := make([]string, len(uniqueLocations))
	for i, loc := range uniqueLocations {
		iataCodes[i] = loc.IATA
	}

	// Fetch all SkyScanner IDs at once
	skyscannerIDs, err := FetchAllSkyScannerIDs(iataCodes)
	if err != nil {
		return nil, err
	}

	// Process each location
	for i, loc := range uniqueLocations {
		// Generate URLs for flights and hotels
		airbnbURL := GenerateAirbnbURL(loc.CityName)
		bookingURL := GenerateBookingURL(loc.CityName)

		// Calculate avg WPI
		temp_five_day_wpi, _ := ProcessLocation(loc, records)
		loc.FiveDayWPI = temp_five_day_wpi
		loc.AirbnbURL = airbnbURL
		loc.BookingURL = bookingURL

		// Fetch skyscanner ID from the map
		loc.SkyScannerID = skyscannerIDs[loc.IATA]

		uniqueLocations[i] = loc
	}

	return uniqueLocations, nil
}

