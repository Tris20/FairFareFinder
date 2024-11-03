
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type selectorInfo struct {
	File  string
	Line  int
	Block string
}

func main() {
	// Get the list of CSS files as input from the command line
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <file1.css> <file2.css> ...")
		return
	}
	files := os.Args[1:]

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
