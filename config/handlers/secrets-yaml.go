package config_handlers

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// Secrets represents the structure of the secrets.yaml file.
type Secrets struct {
	APIKeys map[string]string `yaml:"api_keys"`
}

// loadApiKey loads the API key for a given domain from a YAML file
func LoadApiKey(filePath, domain string) (string, error) {
	var secrets Secrets

	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	err = yaml.Unmarshal(yamlFile, &secrets)
	if err != nil {
		return "", err
	}

	apiKey, ok := secrets.APIKeys[domain]
	if !ok {
		return "", fmt.Errorf("API key for %s not found", domain)
	}

	return apiKey, nil
}
