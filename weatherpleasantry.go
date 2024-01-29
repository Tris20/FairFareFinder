
package main

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

// WeatherPleasantnessConfig holds the configuration for weather pleasantness ratings.
type WeatherPleasantnessConfig struct {
	Conditions map[string]float64 `yaml:"conditions"`
}

// LoadWeatherPleasantnessConfig loads the weather pleasantness configuration from a YAML file.
func LoadWeatherPleasantnessConfig(filePath string) (WeatherPleasantnessConfig, error) {
	var config WeatherPleasantnessConfig

	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

// weatherPleasantness calculates the "weather pleasentness index" (WPI)
func weatherPleasantness(temp float64, wind float64, cond string, config WeatherPleasantnessConfig) float64 {
	weightTemp := 3.0
	weightWind := 1.0
	weightCond := 2.0

	tempindex := tempPleasantness(temp) * weightTemp
	windindex := windPleasantness(wind) * weightWind
	weathindex := weatherCondPleasantness(cond, config) * weightCond

	index := (tempindex + windindex + weathindex) / (weightTemp + weightWind + weightCond)
	return index
}

// tempPleasantness returns a value between 0 and 10 for temperature pleasantness
func tempPleasantness(temperature float64) float64 {
	GoodTemp := 20.0
	indexAtGoodTemp := 7.0
	PerfectTemp := 23.0
	slope := indexAtGoodTemp / GoodTemp

	if temperature <= 0 {
		return 0
	} else if temperature > PerfectTemp {
		return 10
	} else {
		return slope * temperature
	}
}

// weatherCondPleasantness returns a value between 0 and 10 for weather condition pleasantness
func weatherCondPleasantness(cond string, config WeatherPleasantnessConfig) float64 {
	pleasantness, ok := config.Conditions[cond]
	if !ok {
		return 0
	}
	return pleasantness
}

// windPleasantness returns a value between 0 and 10 for wind condition pleasantness
func windPleasantness(windSpeed float64) float64 {
	worstWind := 13.8
	if windSpeed >= worstWind {
		return 0
	} else {
		return 10 - windSpeed*10/worstWind
	}
}
