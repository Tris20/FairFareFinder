package code_analysis

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
)

type selectorInfo struct {
	File  string
	Line  int
	Block string
}

func DetectCSSConflict_FileBased(files []string) {
	fmt.Println("Detecting CSS conflicts based on files")

	// Map to store selectors and their definitions
	selectorMap := make(map[string][]selectorInfo)

	// Regular expression to match CSS selectors (simplified)
	selectorRegex := regexp.MustCompile(`^\s*([^\s{]+)\s*{`)

	for _, file := range files {
		// Open the CSS file
		cssFile, err := os.Open(file)
		if err != nil {
			fmt.Printf("Error opening file %s: %v\n", file, err)
			continue
		}
		defer cssFile.Close()

		scanner := bufio.NewScanner(cssFile)
		lineNum := 0

		// Read the file line by line
		for scanner.Scan() {
			line := scanner.Text()
			lineNum++

			// Check if the line contains a CSS selector
			matches := selectorRegex.FindStringSubmatch(line)
			if matches != nil {
				selector := strings.TrimSpace(matches[1])
				// Store the selector information (file, line number)
				info := selectorInfo{
					File:  filepath.Base(file),
					Line:  lineNum,
					Block: line,
				}
				selectorMap[selector] = append(selectorMap[selector], info)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading file %s: %v\n", file, err)
		}
	}

	// Report conflicts: selectors with more than one definition
	fmt.Println("Conflicting CSS Selectors:")
	for selector, infos := range selectorMap {
		if len(infos) > 1 {
			fmt.Printf("Selector '%s' is defined in multiple places:\n", selector)
			for _, info := range infos {
				fmt.Printf("  - %s:%d (%s)\n", info.File, info.Line, info.Block)
			}
		}
	}
}

// the server needs to be running before callingthis file
func DetectCSSConflict_URLBased() {
	url := "http://localhost:8080/table_view" // Example: change to your URL
	err := detectCSSConflicts(url)
	if err != nil {
		log.Fatalf("Error detecting CSS conflicts: %v", err)
	}
}

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

// func main() {
// 	// URL of the page to analyze
// 	url := "http://localhost:8080/table_view" // Example: change to your URL

// 	// Run the conflict detection
// 	err := detectCSSConflicts(url)
// 	if err != nil {
// 		log.Fatalf("Error detecting CSS conflicts: %v", err)
// 	}
// }
