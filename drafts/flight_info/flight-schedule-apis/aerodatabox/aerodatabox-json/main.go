package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Struct to match the YAML structure
type Secrets struct {
	APIKeys struct {
		Aerodatabox string `yaml:"aerodatabox"`
	} `yaml:"api_keys"`
}

func readAPIKey(filepath string) (string, error) {
	var secrets Secrets
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	err = yaml.Unmarshal(file, &secrets)
	if err != nil {
		return "", err
	}
	return secrets.APIKeys.Aerodatabox, nil
}

func fetchFlightData(url, apiKey string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("X-RapidAPI-Key", apiKey)
	req.Header.Add("X-RapidAPI-Host", "aerodatabox.p.rapidapi.com")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func main() {
	var (
		direction = flag.String("direction", "Departure", "Flight direction: Departure or Arrival")
		airport   = flag.String("airport", "EDI", "IATA airport code")
		date      = flag.String("date", "27-02-2024", "Date in DD-MM-YYYY format")
	)
	flag.Parse()

	apiKey, err := readAPIKey("../../../../../ignore/secrets.yaml")
	if err != nil {
		fmt.Println("Error reading API key:", err)
		return
	}

	// Ensure the results directory exists
	resultsDir := "./results"
	if _, err := os.Stat(resultsDir); os.IsNotExist(err) {
		err = os.Mkdir(resultsDir, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating results directory:", err)
			return
		}
	}

	dateParts := strings.Split(*date, "-")
	if len(dateParts) != 3 {
		fmt.Println("Invalid date format. Please use DD-MM-YYYY.")
		return
	}
	dateFormatted := fmt.Sprintf("%s-%s-%s", dateParts[2], dateParts[1], dateParts[0])

	directionPrefix := "DEP"
	if *direction == "Arrival" {
		directionPrefix = "ARR"
	}

	intervals := []struct {
		urlSuffix  string
		fileSuffix string
	}{
		{"T00:00/%sT11:59", "AM"},
		{"T12:00/%sT23:59", "PM"},
	}

	for _, interval := range intervals {
		url := fmt.Sprintf("https://aerodatabox.p.rapidapi.com/flights/airports/iata/%s/%s"+interval.urlSuffix+"?withLeg=true&direction=%s&withCancelled=true&withCodeshared=true&withLocation=false", *airport, dateFormatted, dateFormatted, *direction)
		body, err := fetchFlightData(url, apiKey)
		if err != nil {
			fmt.Println("Error fetching data:", err)
			return
		}
		// Corrected filename formatting to avoid EXTRA string issue
		fileName := fmt.Sprintf("%s/%s-%s-%s-%s-%s-%s.json", resultsDir, *airport, directionPrefix, dateParts[0], dateParts[1], dateParts[2], interval.fileSuffix)
		err = os.WriteFile(fileName, []byte(body), 0644)
		if err != nil {
			fmt.Println("Error writing file:", err)
			return
		}
		fmt.Printf("Data saved to %s\n", fileName)
	}

}
