// model/types.go
package model

type WeatherData struct {
	Dt   int64 `json:"dt"` // Unix timestamp of the forecasted data
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Weather []struct {
		Main string `json:"main"`
	} `json:"weather"`
}