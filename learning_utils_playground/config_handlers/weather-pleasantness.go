package config_handlers

// Configs and input handlers go here

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type WeatherPleasantnessConfig struct {
	Conditions map[string]float64 `yaml:"conditions"`
}

func LoadWeatherPleasantnessConfig(filePath string) (WeatherPleasantnessConfig, error) {

	var config WeatherPleasantnessConfig
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(yamlFile, &config)
	return config, err
}
