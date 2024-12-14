package main

// because the file uses the same package and is in the same directory as main.go
// it can access the functions and variables in main.go that are private

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
)

// TestMain is a test setup function. It is called before any tests are run.
// It is used to set up any resources that are needed by the tests.
// This is a feature of the testing package.
func TestMain(m *testing.M) {
	// setup resources / set up
	SetupServer("./testdata/test.db")
	// Run the test
	exitVal := m.Run()

	// cleanup resources / additional tear down

	// exit
	os.Exit(exitVal)
}

func TestFilterHandler(t *testing.T) {
	// run the server for this test
	go StartServer()
	// wait for the server to start
	time.Sleep(10 * time.Millisecond)

	// Send the request
	url := "http://127.0.0.1:8080/filter?city[]=Berlin&maxPriceLinear[]=14"
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

	// Define the expected destination
	expectedDestination := "Tivat"
	// Check the first destination
	firstRow := doc.Find("tbody > tr").First()
	firstDestination := firstRow.Find("td").First().Text()
	firstDestination = strings.TrimSpace(firstDestination)
	if expectedDestination != firstDestination {
		t.Errorf("Expected '%s' to be the first destination, but got '%s'.", expectedDestination, firstDestination)
	}

	// Check the number of weather columns in the first row
	weatherColumns := firstRow.Find("td.weather-column")
	if weatherColumns.Length() != 5 {
		t.Errorf("Expected 5 weather columns for '%s', but found %d.", expectedDestination, weatherColumns.Length())
	}
}

// test for combinedCardsHandler function
func TestCombinedCardsHandler(t *testing.T) {
	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/filter", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add query parameters to the request
	q := req.URL.Query()
	q.Add("city[]", "Berlin")
	q.Add("maxPriceLinear[]", "200")
	q.Add("maxAccommodationPrice[]", "150")
	req.URL.RawQuery = q.Encode()

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function
	handler := http.HandlerFunc(combinedCardsHandler)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body, just look for something from table.html
	expected := `<th class="fnaf-column">Five Nights and Flights<br />(Per Person)</th>`
	if !strings.Contains(rr.Body.String(), expected) {
		if len(rr.Body.String()) < 1000 {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		} else {
			t.Errorf("handler returned unexpected body: size too large to show, but want %v",
				expected)
		}
	}
}

// test for combinedCardsHandler function with invalid input lengths
func TestCombinedCardsHandler_InvalidInputLengths(t *testing.T) {
	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/filter", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add query parameters to the request
	q := req.URL.Query()
	q.Add("city[]", "New York")
	q.Add("logical_operator[]", "AND")
	q.Add("maxPriceLinear[]", "100")
	req.URL.RawQuery = q.Encode()

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function
	handler := http.HandlerFunc(combinedCardsHandler)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	// Check the response body
	expected := "Mismatched input lengths"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

// test for combinedCardsHandler function with invalid price parameter
func TestCombinedCardsHandler_InvalidPriceParameter(t *testing.T) {
	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/filter", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add query parameters to the request
	q := req.URL.Query()
	q.Add("city[]", "New York")
	q.Add("city[]", "Los Angeles")
	q.Add("logical_operator[]", "AND")
	q.Add("maxPriceLinear[]", "invalid")
	q.Add("maxPriceLinear[]", "200")
	q.Add("maxAccommodationPrice[]", "150")
	req.URL.RawQuery = q.Encode()

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function
	handler := http.HandlerFunc(combinedCardsHandler)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	// Check the response body
	expected := "Invalid price parameter"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

// test for combinedCardsHandler function with invalid accommodation price parameter
func TestCombinedCardsHandler_InvalidAccommodationPriceParameter(t *testing.T) {
	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/filter", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add query parameters to the request
	q := req.URL.Query()
	q.Add("city[]", "New York")
	q.Add("city[]", "Los Angeles")
	q.Add("logical_operator[]", "AND")
	q.Add("maxPriceLinear[]", "100")
	q.Add("maxPriceLinear[]", "200")
	q.Add("maxAccommodationPrice[]", "invalid")
	req.URL.RawQuery = q.Encode()

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function
	handler := http.HandlerFunc(combinedCardsHandler)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	// Check the response body
	expected := "Invalid accommodation price parameter"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
