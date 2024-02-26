package main

import (
	"fmt"
	"net/http"
	"io"
)

func main() {

	url := "https://timetable-lookup.p.rapidapi.com/airports/BER/routes/nonstops/"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("X-RapidAPI-Key", "APIKEY")
	req.Header.Add("X-RapidAPI-Host", "timetable-lookup.p.rapidapi.com")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))

}
