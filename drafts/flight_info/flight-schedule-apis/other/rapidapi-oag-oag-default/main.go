package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {

	url := "https://flight-info-api.p.rapidapi.com/schedules?version=v2&DepartureDateTime=2024-02-28%2F2024-03-02&DepartureAirport=BER&CodeType=IATA&ServiceType=Passenger&limit=1600"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("X-RapidAPI-Key", "APIKEY")
	req.Header.Add("X-RapidAPI-Host", "flight-info-api.p.rapidapi.com")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))

}
