package config_handlers

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type OriginInfo struct {
	IATA               string `yaml:"IATA"`
	City               string `yaml:"City"`
	Country            string `yaml:"Country"`
	DepartureStartDate string `yaml:"DepartureStartDate"`
	DepartureEndDate   string `yaml:"DepartureEndDate"`
	ArrivalStartDate   string `yaml:"ArrivalStartDate"`
	ArrivalEndDate     string `yaml:"ArrivalEndDate"`
	SkyScannerID       string `yaml:"SkyScannerID"`
}

type Config struct {
	Origins []OriginInfo `yaml:"origins"`
}

func LoadOrigins(filename string) ([]OriginInfo, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return nil, err
	}

	return config.Origins, nil
}
