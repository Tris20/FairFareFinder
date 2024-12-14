package main

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery" // Use a robust HTML parser
)

func TestFilterHandler(t *testing.T) {
	// Define the expected destination
	expectedDestination := "Tivat"

	// Send the request
	url := "http://127.0.0.1:8080/filter?city1=Berlin&maxPriceLinear=14"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read response body: %v", err)
		} else {
			bodyString := string(bodyBytes)
			t.Errorf("Response body: %s", bodyString)
		}
	}

	// Parse the response body using goquery for structured HTML parsing
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	// Check the first destination
	firstRow := doc.Find("tbody > tr").First()
	firstDestination := firstRow.Find("td").First().Text()
	if strings.TrimSpace(firstDestination) != expectedDestination {
		t.Errorf("Expected '%s' to be the first destination, but got '%s'.", expectedDestination, firstDestination)
	}

	// Check the number of weather columns in the first row
	weatherColumns := firstRow.Find("td.weather-column")
	if weatherColumns.Length() != 5 {
		t.Errorf("Expected 5 weather columns for '%s', but found %d.", expectedDestination, weatherColumns.Length())
	}
}
