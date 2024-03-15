package discourse

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Config struct adjusted for the nested api_keys
type config struct {
	ApiKeys map[string]string `yaml:"api_keys"`
}

// readApiKey reads the API key for partybus from secrets.yaml
func readApiKey() (string, error) {

	filePath := "./ignore/secrets.yaml" // Adjust the path as needed
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var cfg config
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return "", err
	}

	apiKey, ok := cfg.ApiKeys["partybus"]
	if !ok {
		return "", fmt.Errorf("API key for partybus not found")
	}

	return apiKey, nil
}

// PostToDiscourse is an exported function to post content to a Discourse post
func PostToDiscourse(content string) error {
	apiKey, err := readApiKey()
	if err != nil {
		return err
	}

	apiUsername := "tristan"
	postURL := "https://partybus.community/posts/392.json"
	client := &http.Client{}
	data := url.Values{}
	data.Set("post[raw]", content)
	fmt.Println("\nPOSTING")
	req, err := http.NewRequest("PUT", postURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Api-Key", apiKey)
	req.Header.Set("Api-Username", apiUsername)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Response status:", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println("Response body:", string(body))
	return nil
}
