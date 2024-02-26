
package main

import (
	"encoding/json"
	"io/ioutil"
)

// ParseJSON parses the JSON file containing flight data
func ParseJSON(filepath string) (*APIResponse, error) {
    data, err := ioutil.ReadFile(filepath)
    if err != nil {
        return nil, err // Return nil and the error
    }

    var apiResponse APIResponse
    err = json.Unmarshal(data, &apiResponse)
    if err != nil {
        return nil, err // Return nil and the error
    }

    return &apiResponse, nil // Return the parsed data and nil as the error
}
