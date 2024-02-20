
package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url" // Added for url.QueryEscape
    "os"
    "path/filepath"
    "strings"
    "encoding/json"
    "gopkg.in/yaml.v2"
)

// Config structure to match the YAML
type Config struct {
    ApiKeys struct {
        BingAI string `yaml:"bingai"`
    } `yaml:"api_keys"`
}


type BingSearchResponse struct {
    WebPages struct {
        WebSearchURL        string `json:"webSearchUrl"`
        TotalEstimatedMatches int    `json:"totalEstimatedMatches"`
        Value               []struct {
            Name                  string `json:"name"`
            URL                   string `json:"url"`
            Snippet               string `json:"snippet"`
            DatePublished         string `json:"datePublished"`
            DatePublishedFreshnessText string `json:"datePublishedFreshnessText"`
            DisplayURL            string `json:"displayUrl"`
        } `json:"value"`
    } `json:"webPages"`
}

// Function to read the API key from YAML
func readAPIKeyFromYAML(filePath string) (string, error) {
    var config Config
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return "", err
    }
    if err := yaml.Unmarshal(data, &config); err != nil {
        return "", err
    }
    return config.ApiKeys.BingAI, nil
}

// Function to query Bing Search API
func queryBing(apiKey, query string) (string, error) {
    endpoint := fmt.Sprintf("https://api.bing.microsoft.com/v7.0/search?q=%s", url.QueryEscape(query))
    req, _ := http.NewRequest("GET", endpoint, nil)
    req.Header.Add("Ocp-Apim-Subscription-Key", apiKey)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(body), nil // Now we're returning the body correctly
}



// Function to generate HTML content
func generateHTML(cities []string, apiKey string) (string, error) {
    var sb strings.Builder
    sb.WriteString("<!DOCTYPE html>\n<html>\n<head>\n<title>City Highlights</title>\n</head>\n<body>\n")
    sb.WriteString("<table border='1'>\n<tr><th>City</th><th>Highlights</th></tr>\n")

    for _, city := range cities {
        query := fmt.Sprintf("Things to do in %s this weekend", city)
        responseJSON, err := queryBing(apiKey, query)
        if err != nil {
            fmt.Printf("Error querying Bing for %s: %v\n", city, err)
            continue
        }

        var response BingSearchResponse
        json.Unmarshal([]byte(responseJSON), &response)

        sb.WriteString(fmt.Sprintf("<tr><td>%s</td><td>", city))
        for _, item := range response.WebPages.Value {
            sb.WriteString(fmt.Sprintf("<a href='%s'>%s</a><br>%s<br><br>", item.URL, item.Name, item.Snippet))
        }
        sb.WriteString("</td></tr>\n")
    }

    sb.WriteString("</table>\n</body>\n</html>")
    return sb.String(), nil
}

func main() {
    cities := []string{"New York", "San Francisco", "Paris", "Tokyo"}
    yamlPath := filepath.Join("..", "..", "ignore", "secrets.yaml")
    apiKey, err := readAPIKeyFromYAML(yamlPath)
    if err != nil {
        fmt.Println("Error reading API key:", err)
        os.Exit(1)
    }

    htmlContent, err := generateHTML(cities, apiKey)
    if err != nil {
        fmt.Println("Error generating HTML:", err)
        os.Exit(1)
    }

    // Write the HTML content to a file
    if err := ioutil.WriteFile("city_highlights.html", []byte(htmlContent), 0644); err != nil {
        fmt.Println("Error writing HTML file:", err)
        os.Exit(1)
    }

    fmt.Println("HTML file generated successfully.")
}

