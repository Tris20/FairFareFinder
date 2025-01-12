package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Result holds information about duplicate definitions.
type Result struct {
	File      string
	Line      int
	Class     string
	Attribute string
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run main.go <directory> <output_file> <folder_depth>")
		return
	}
	directory := os.Args[1]
	outputFile := os.Args[2]
	folderDepth := parseDepthArg(os.Args[3])

	results := []Result{}
	duplicates := map[string]map[string][]Result{} // Class -> Attribute -> Results

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".css") {
			relPath := getLastNParts(path, folderDepth)
			fileResults, err := processFile(path, relPath)
			if err != nil {
				fmt.Printf("Error processing file %s: %v\n", path, err)
				return nil
			}
			for _, res := range fileResults {
				if _, ok := duplicates[res.Class]; !ok {
					duplicates[res.Class] = map[string][]Result{}
				}
				if _, exists := duplicates[res.Class][res.Attribute]; !exists {
					duplicates[res.Class][res.Attribute] = []Result{}
				}
				duplicates[res.Class][res.Attribute] = append(duplicates[res.Class][res.Attribute], res)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return
	}

	for class, attrs := range duplicates {
		for attr, resList := range attrs {
			if len(resList) > 1 {
				results = append(results, resList...)
			}
		}
	}

	if err := writeResults(outputFile, results); err != nil {
		fmt.Printf("Error writing results: %v\n", err)
	}
}

func parseDepthArg(arg string) int {
	depth, err := strconv.Atoi(arg)
	if err != nil || depth <= 0 {
		fmt.Println("Invalid folder depth, defaulting to 3")
		return 3
	}
	return depth
}

func getLastNParts(path string, depth int) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) > depth {
		return strings.Join(parts[len(parts)-depth:], "/")
	}
	return strings.Join(parts, "/")
}

func processFile(fullPath string, relPath string) ([]Result, error) {
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	classPattern := regexp.MustCompile(`(?m)^\s*(\.\w[\w\-]*)\s*\{`)
	attrPattern := regexp.MustCompile(`(?m)^\s*([a-zA-Z-]+)\s*:\s*[^;]+;`)

	results := []Result{}
	currentClass := ""
	lineNumber := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++

		// Match class/element
		if matches := classPattern.FindStringSubmatch(line); matches != nil {
			currentClass = matches[1]
			continue
		}

		// Match attributes if in a class
		if currentClass != "" {
			if matches := attrPattern.FindStringSubmatch(line); matches != nil {
				results = append(results, Result{
					File:      relPath,
					Line:      lineNumber,
					Class:     currentClass,
					Attribute: matches[1],
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func writeResults(outputFile string, results []Result) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, res := range results {
		_, err := writer.WriteString(fmt.Sprintf("%s:%d\t%s\t%s\n",
			res.File, res.Line, res.Class, res.Attribute))
		if err != nil {
			return err
		}
	}

	return nil
}
