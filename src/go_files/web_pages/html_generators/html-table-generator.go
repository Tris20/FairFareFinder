package htmltablegenerator

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Tris20/FairFareFinder/src/go_files"
	"github.com/Tris20/FairFareFinder/src/go_files/timeutils"
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

var number_of_day_columns int
var daycolumn_min int
var daycolumn_max int

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
    <link rel="stylesheet" href="../tableStyles.css">
<link rel="shortcut icon" href="/images/favicon.ico" type="image/x-icon">
</head>
<body>
<table>
<thead>
<tr>
    <th>City Name</th>`)

	// Determine day order based on current day
	daysOrder, startDay, endDay := timeutils.GetDaysOrder()

	// Map to store slices of DailyWeatherDetails by Weekday for easy lookup
	dailyDetailsByDay := make(map[time.Weekday][]model.DailyWeatherDetails)
	for _, detail := range citiesData[0].WeatherDetails {
		dailyDetailsByDay[detail.Day] = append(dailyDetailsByDay[detail.Day], detail)
	}

	number_of_day_columns = 0
	daycolumn_min = 0
	daycolumn_max = 0
	// Iterate over the daysOrder slice to maintain order
	for _, dayOfWeek := range daysOrder {
		if timeutils.ShouldIncludeDay(dayOfWeek, startDay, endDay) {
			if _, ok := dailyDetailsByDay[dayOfWeek]; ok {
				//	for _, dayDetail := range details {

				//	for _, dayOfWeek := range daysOrder {
				dayhtml := fmt.Sprintf(`<th style="width: 70px; ">%s</th>`, dayOfWeek)
				_, err = writer.WriteString(dayhtml)
				number_of_day_columns += 1
				daycolumn_max += 1
			} else {
				daycolumn_min += 1
			}
		}
	}
	_, err = writer.WriteString(`<th>Flights</th>
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

	daysOrder, startDay, endDay := timeutils.GetDaysOrder()

	// Map to store slices of DailyWeatherDetails by Weekday for easy lookup
	dailyDetailsByDay := make(map[time.Weekday][]model.DailyWeatherDetails)
	for _, detail := range destination.WeatherDetails {
		dailyDetailsByDay[detail.Day] = append(dailyDetailsByDay[detail.Day], detail)
	}

	// Compile the regular expression outside of the loop to optimize performance
	iconFormat := regexp.MustCompile(`^\d{2}[a-z]$`)
	fmt.Println("Days Order:", daysOrder)
	for day, details := range dailyDetailsByDay {
		fmt.Printf("Day: %v, Details: %+v\n", day, details)
	}
	// Iterate over the daysOrder slice to maintain order
	for day_number, dayOfWeek := range daysOrder {
		if timeutils.ShouldIncludeDay(dayOfWeek, startDay, endDay) {

			if details, ok := dailyDetailsByDay[dayOfWeek]; ok {
				for _, dayDetail := range details {
					// Check if the icon format is valid
					if iconFormat.MatchString(dayDetail.Icon) {
						//convert temp to string because sprintf or writestring struggled with floats
						avg_temp := fmt.Sprintf("%0.1f°C", dayDetail.AverageTemp)
						weatherHTML.WriteString(fmt.Sprintf(
							`<td ><a href="https://www.google.com/search?q=weather+%s"><img src="http://openweathermap.org/img/wn/%s.png" alt="Weather Icon" style="max-width:100%%; height:auto;" ></a> <br><span>%s</span></td>`, destination.City, dayDetail.Icon, avg_temp))
					} else {

						// Invalid icon format - replace with a default icon or just a hyperlink
						// Assuming "default.png" is your default icon. Adjust the src attribute as needed.
						if daycolumn_min <= day_number && day_number <= daycolumn_max {

							weatherHTML.WriteString(fmt.Sprintf(
								`<td><a href="https://www.google.com/search?q=weather+%s"><img src="src/images/unknownweather.png" alt="Default Weather Icon" style="max-width:100%%; height:auto;"></a></td> `, destination.City))
						}
					}
				}
			} else {
				if daycolumn_min <= day_number && day_number <= daycolumn_max {
					weatherHTML.WriteString(fmt.Sprintf(
						`<td><a href="https://www.google.com/search?q=weather+%s"><img src="/images/unknownweather.png" alt="Default Weather Icon" style="max-width:100%%; height:auto;"></a></td> `, destination.City))
				}
			}
		}
	}
	skyscannertext := "SkyScanner"
	if destination.SkyScannerPrice > 0.0 {
		skyscannertext = fmt.Sprintf("From €%.2f", destination.SkyScannerPrice)
	}

	return fmt.Sprintf(
		`<tr>
    <td><a href="https://www.google.com/maps/place/%[1]s">%[1]s</a></td>
    %s
    <td><a href="%s">%s</a></td>
    <td><a href="%s">Airbnb</a> <a href="%s">Booking.com</a></td>
    <td><a href="https://www.google.com/search?q=things+to+do+this+weekend+%s">Google Results</a></td>
</tr>`, destination.City, weatherHTML.String(), destination.SkyScannerURL, skyscannertext, destination.AirbnbURL, destination.BookingURL, destination.City)
}
