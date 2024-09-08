package config_handlers

import (
	"github.com/Tris20/FairFareFinder/src/backend"
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

// ConvertConfigToModel converts a slice of config_handlers.OriginInfo to a slice of model.OriginInfo
func ConvertConfigToModel(originsConfig []OriginInfo) []model.OriginInfo {
	var originsModel []model.OriginInfo
	for _, originConfig := range originsConfig {
		originModel := model.OriginInfo{
			IATA:               originConfig.IATA,
			City:               originConfig.City,
			Country:            originConfig.Country,
			DepartureStartDate: originConfig.DepartureStartDate,
			DepartureEndDate:   originConfig.DepartureEndDate,
			ArrivalStartDate:   originConfig.ArrivalStartDate,
			ArrivalEndDate:     originConfig.ArrivalEndDate,
			SkyScannerID:       originConfig.SkyScannerID,
		}
		originsModel = append(originsModel, originModel)
	}
	return originsModel
}
