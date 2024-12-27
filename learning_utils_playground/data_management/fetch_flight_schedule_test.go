package data_management

import (
	"bufio"
	"bytes"
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

type MockAPIClient struct {
	apiKey  string
	logData []string
	index   int
}

func NewMockAPIClient(logFilePath string) (*MockAPIClient, error) {
	client := &MockAPIClient{}

	file, err := os.Open(logFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var buffer bytes.Buffer
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if strings.HasPrefix(line, "Response: ") {
			if buffer.Len() > 0 {
				client.logData = append(client.logData, buffer.String())
				buffer.Reset()
			}
			response := strings.TrimPrefix(line, "Response: ")
			buffer.WriteString(response)
		}
	}
	if buffer.Len() > 0 {
		client.logData = append(client.logData, buffer.String())
	}

	return client, nil
}

func (c *MockAPIClient) FetchFlightData(url string) ([]byte, error) {

	if len(c.logData) == 0 {
		return nil, nil
	}

	// loop over the data once
	for i := 0; i < len(c.logData); i++ {
		response := c.logData[c.index]
		c.index = (c.index + 1) % len(c.logData)

		if strings.Contains(url, "direction=Arrival") {
			if strings.Contains(response[0:30], `{"arrivals":`) {
				return []byte(response), nil
			}
		} else if strings.Contains(url, "direction=Departure") {
			if strings.Contains(response[0:30], `{"departures":`) {
				return []byte(response), nil
			}
		}
	}
	// return empty response if no match found
	return []byte(""), nil
}

func (c *MockAPIClient) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
}

func TestFetchFlightSchedule(t *testing.T) {
	// fmt.Println("Setting up mock api client")
	mockClient, err := NewMockAPIClient("testdata/real_flight_api_responses.test_log")
	if err != nil {
		t.Fatalf("Failed to create MockAPIClient: %v", err)
	}
	mockClient.SetAPIKey("test")
	configFilePath := "../../config/config.yaml"
	secretsFilePath := "../../ignore/secrets.yaml"
	flightsDBPath := "testdata/flights.db"

	// remove the database if it exists
	os.Remove(flightsDBPath)

	// fmt.Println("Running FetchFlightSchedule")
	// function should run without error
	err = FetchFlightSchedule(mockClient, configFilePath, secretsFilePath, flightsDBPath)
	if err != nil {
		t.Fatalf("FetchFlightSchedule returned an error: %v", err)
	}

	// check that the database was created
	db, err := sql.Open("sqlite3", flightsDBPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
}

// WARNING: this test is not really a test. it makes real api calls and logs the responses
// this is used to generate test data
//
// uncomment and run just this test to make new data. It will time out after 30 seconds

// func TestProduceRealAPIResponses(t *testing.T) {
// 	configFilePath := "../../config/config.yaml"
// 	secretsFilePath := "../../ignore/secrets.yaml"
// 	flightsDBPath := "testdata/flights.db"

// 	// Set up the real client to log responses
// 	realClient := NewRealFlightAPIClient()
// 	realClient.SetAPIKey("real_api_key")
// 	err := realClient.SetLogFile("testdata/real_flight_api_responses.test_log")
// 	if err != nil {
// 		t.Fatalf("Failed to set log file: %v", err)
// 	}

// 	err = FetchFlightSchedule(realClient, configFilePath, secretsFilePath, flightsDBPath)
// 	if err != nil {
// 		t.Fatalf("FetchFlightSchedule returned an error: %v", err)
// 	}

// 	// Add assertions to verify the behavior
// }

func TestFetchFlightSchedule2(t *testing.T) {
	// FetchFlightSchedule("arrivals")
	// FetchFlightSchedule("departures")
}

func TestGetFileDependencies(t *testing.T) {
	// Create temporary config and secrets files
	configFilePath := "testdata/test_config.yaml"
	secretsFilePath := "testdata/test_secrets.yaml"
	flightsDBPath := "testdata/test_flights.db"

	// 	configContent := `
	// airports:
	//   - "BER"
	//   - "GLA"
	//   - "EDI"
	//   - "MUC"
	//   - "FRA"
	//   - "SYD"
	// `
	// 	secretsContent := `
	// api_keys: test_api_key
	// aerodatabox: test_aerodatabox_key
	// `

	// err := os.WriteFile(configFilePath, []byte(configContent), 0644)
	// if err != nil {
	// 	t.Fatalf("Failed to write config file: %v", err)
	// }
	// defer os.Remove(configFilePath)

	// err = os.WriteFile(secretsFilePath, []byte(secretsContent), 0644)
	// if err != nil {
	// 	t.Fatalf("Failed to write secrets file: %v", err)
	// }
	// defer os.Remove(secretsFilePath)

	// Create a temporary SQLite database
	db, err := sql.Open("sqlite3", flightsDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer os.Remove(flightsDBPath)
	defer db.Close()

	// Call the function
	configs, apiKey, db, err := getFileDependencies(configFilePath, secretsFilePath, flightsDBPath)
	if err != nil {
		t.Fatalf("getFileDependencies returned an error: %v", err)
	}

	// Verify the results
	if len(configs.Airports) == 0 || configs.Airports[0] != "BER" {
		t.Errorf("Unexpected configs: %+v", configs)
	}

	if apiKey != "test_aerodatabox_key" {
		t.Errorf("Unexpected apiKey: %s", apiKey)
	}

	if db == nil {
		t.Errorf("Expected db to be non-nil")
	}
}
