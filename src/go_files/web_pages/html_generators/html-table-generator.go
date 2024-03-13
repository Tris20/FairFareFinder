package htmltablegenerator

import (
	"bufio"
	"fmt"
	"github.com/Tris20/FairFareFinder/src/go_files"
	"os"
	"strings"
)

// Define a struct for the weather information of a day.
type WeatherDay struct {
	Day  string
	Icon string
}

// Define a struct for each city's data, including weather forecast and links.
type CityData struct {
	Name              string
	WeatherForecast   []WeatherDay
	FlightLink        string
	AccommodationLink string
	ThingsToDoLink    string
}

// GenerateHtmlTable creates an HTML table for multiple cities and saves it to the specified file path.
func GenerateHtmlTable(outputPath string, citiesData []model.DestinationInfo) error {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)

	// Start of the HTML structure
	_, err = writer.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>City Weather Forecast Table</title>
</head>
<body>
<table>
<thead>
<tr>
    <th>City Name</th>
    <th>WPI</th>
    <th>Flights</th>
    <th>Accommodation</th>
    <th>Things to Do</th>
</tr>
</thead>
<tbody>
`)
	if err != nil {
		return fmt.Errorf("error writing header to output file: %w", err)
	}

	// Generate and write the HTML for each city's table row
	for _, city := range citiesData {
		tableRow := generateTableRow(city)
		if _, err := writer.WriteString(tableRow); err != nil {
			return fmt.Errorf("error writing table row to output file: %w", err)
		}
	}

	// End of the HTML structure
	_, err = writer.WriteString("</tbody>\n</table>\n</body>\n</html>")
	if err != nil {
		return fmt.Errorf("error writing footer to output file: %w", err)
	}

	// Flush the buffer to ensure all data is written to the file
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing to output file: %w", err)
	}

	return nil // No error occurred
}

// generateTableRow generates an HTML table row for the given city data.
func generateTableRow(destination model.DestinationInfo) string {
	var weatherHTML strings.Builder

	for _, day := range destination.WeatherDetails {
		weatherHTML.WriteString(fmt.Sprintf(
			`<span style="display: inline-block; text-align: center; width: 100px;">%s<br><a href="https://www.google.com/search?q=weather+%s"><img src="http://openweathermap.org/img/wn/%s.png" alt="Image" style="max-width:100px;"></a></span> `,
			day.Day, destination.City, day.Icon))
	}

	return fmt.Sprintf(
		`<tr>
    <td><a href="https://www.google.com/maps/place/%[1]s">%[1]s</a></td>
    <td style="white-space: nowrap;">%s</td>
    <td><a href="%s">SkyScanner</a></td>
    <td><a href="%s">Airbnb</a> <a href="%s">Booking.com</a></td>
    <td><a href="https://www.google.com/search?q=things+to+do+this+weekend+%s">Google Results</a></td>
    <td></td>
</tr>`, destination.City, weatherHTML.String(), destination.SkyScannerURL, destination.AirbnbURL, destination.BookingURL, destination.City)
}
