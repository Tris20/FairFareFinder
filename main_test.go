package main

import (
	"database/sql"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestMain is a test setup function. It is called before any tests are run.
// It is used to set up any resources that are needed by the tests.
// This is a feature of the testing package.
func TestMain(m *testing.M) {
	// setup resources / set up
	var err error
	db, err = setupTestDB()
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}
	defer db.Close()

	// Run the test
	exitVal := m.Run()

	// cleanup resources / additional tear down

	// exit
	os.Exit(exitVal)
}

func setupTestDB() (*sql.DB, error) {
	// We can keep a database file for testing purposes
	// From the go book: testdata is a special directory that is reserved by the toolchain
	// path must be relative to the package directory
	testDB, err := sql.Open("sqlite3", "./testdata/test.db")
	return testDB, err
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
	q.Add("city[]", "New York")
	q.Add("city[]", "Los Angeles")
	q.Add("logical_operator[]", "AND")
	q.Add("maxPriceLinear[]", "100")
	q.Add("maxPriceLinear[]", "200")
	q.Add("maxAccommodationPrice[]", "150")
	req.URL.RawQuery = q.Encode()

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function
	handler := http.HandlerFunc(combinedCardsHandler)
	handler.ServeHTTP(rr, req)

	// Check the status code
	// if status := rr.Code; status != http.StatusOK {
	// 	t.Errorf("handler returned wrong status code: got %v want %v",
	// 		status, http.StatusOK)
	// }

	// Check the response body
	// expected := "Generated SQL Query:"
	// if !strings.Contains(rr.Body.String(), expected) {
	// 	t.Errorf("handler returned unexpected body: got %v want %v",
	// 		rr.Body.String(), expected)
	// }
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
