package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
)

// Download CSS from the browser's final state after animations
func fetchFinalCSS(url string) (string, error) {
	// Create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Variable to hold the final CSS
	var cssContent string

	// Run chromedp to navigate the page, wait for animations, and fetch the computed CSS
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		// Wait for 6 seconds to allow animations or JavaScript to load
		chromedp.Sleep(6*time.Second),
		// Retrieve all CSS styles after animations are done
		chromedp.Evaluate(`Array.from(document.styleSheets).map(sheet => {
			try {
				return Array.from(sheet.cssRules).map(rule => rule.cssText).join("\n");
			} catch (e) {
				return ''; // Some stylesheets may be blocked by CORS, skip those
			}
		}).join("\n");`, &cssContent),
	)

	if err != nil {
		return "", fmt.Errorf("failed to retrieve CSS: %v", err)
	}

	return cssContent, nil
}

// Parse the CSS to find conflicting definitions and manually track line numbers
func parseCSS(cssContent string, fileName string, conflicts map[string]map[string][]string, lineNumbers map[string]map[string][]int) {
	cssReader := bytes.NewReader([]byte(cssContent))
	input := parse.NewInput(cssReader)
	lexer := css.NewLexer(input)

	var selector string
	currentLine := 1 // Manually track the line number

	for {
		tt, text := lexer.Next()
		textStr := string(text)

		// Update current line by counting the newlines
		currentLine += strings.Count(textStr, "\n")

		switch tt {
		case css.ErrorToken:
			return
		case css.IdentToken, css.HashToken, css.DelimToken, css.AtKeywordToken:
			selector += textStr

		case css.LeftBraceToken:
			// We've got a selector
			selector = strings.TrimSpace(selector)

		case css.RightBraceToken:
			// Reset selector after closing brace
			selector = ""

		case css.ColonToken:
			// After colon comes the property
			property := selector
			tt, text = lexer.Next()
			// Manually track the line number for the current property
			lineNumber := currentLine

			// Store the conflicts: map[selector][property] = []fileNames
			if _, exists := conflicts[selector]; !exists {
				conflicts[selector] = make(map[string][]string)
				lineNumbers[selector] = make(map[string][]int)
			}
			conflicts[selector][property] = append(conflicts[selector][property], fileName)
			lineNumbers[selector][property] = append(lineNumbers[selector][property], lineNumber)

			// Reset selector for the next rule
			selector = ""
		}
	}
}

// Detect and print CSS conflicts after animations are loaded
func detectCSSConflicts(url string) error {
	// Fetch the final state of the page's CSS after animations
	finalCSS, err := fetchFinalCSS(url)
	if err != nil {
		return err
	}

	// A map to store conflicts: selector -> property -> [files where it's defined]
	conflicts := make(map[string]map[string][]string)
	// A map to store the line numbers: selector -> property -> [line numbers]
	lineNumbers := make(map[string]map[string][]int)

	// Analyze the final CSS state
	fmt.Printf("Analyzing final CSS state from: %s\n", url)
	parseCSS(finalCSS, url, conflicts, lineNumbers)

	// Print the conflicts
	fmt.Println("\nDetected CSS Conflicts:")
	for selector, props := range conflicts {
		for property, files := range props {
			if len(files) > 1 {
				fmt.Printf("Conflict for selector '%s' on property '%s':\n", selector, property)
				for i, file := range files {
					fmt.Printf("  - Defined in: %s at line %d\n", file, lineNumbers[selector][property][i])
				}
			}
		}
	}

	return nil
}

func main() {
	// URL of the page to analyze
	url := "http://localhost:8080/table_view" // Example: change to your URL

	// Run the conflict detection
	err := detectCSSConflicts(url)
	if err != nil {
		log.Fatalf("Error detecting CSS conflicts: %v", err)
	}
}
