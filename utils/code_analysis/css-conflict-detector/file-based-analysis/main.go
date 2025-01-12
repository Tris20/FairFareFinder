package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// CSSInfo holds all the discovered CSS declarations
// key:   selector name (e.g. ".button", "#header", "h1", etc.)
// value: map[propertyName][]ConflictDetail
type CSSInfo map[string]map[string][]ConflictDetail

// ConflictDetail stores info about where and how a property was declared
type ConflictDetail struct {
	FilePath   string
	Line       int
	Value      string
	MediaQuery string // e.g. "@media screen and (max-width: 800px)"
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run main.go <directory> <output-file> <folder-depth>")
		os.Exit(1)
	}

	// 1) Directory to scan for CSS files
	dir := os.Args[1]

	// 2) Output file
	outFile := os.Args[2]

	// 3) Folder depth (how many path segments from the end we want to display)
	folderDepth, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatalf("Invalid folder-depth argument: %v", err)
	}

	// Create or overwrite the output file
	f, err := os.Create(outFile)
	if err != nil {
		log.Fatalf("Failed to create output file %s: %v", outFile, err)
	}
	defer f.Close()

	// We'll store all of our CSS parse results here
	cssInfo := make(CSSInfo)

	// Walk the directory and parse each .css file
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Check if it's a CSS file
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), ".css") {
			parseCSSFile(path, cssInfo)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to walk directory: %v", err)
	}

	// Now collect conflicts
	conflicts := collectConflicts(cssInfo)

	// Print conflicts (or "No conflicts") to the specified output
	printConflicts(conflicts, f, folderDepth)
}

// parseCSSFile scans a single CSS file, finds selectors, and extracts property-value pairs (naïve).
func parseCSSFile(filePath string, cssInfo CSSInfo) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Failed to open CSS file %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Regex to find selectors that start a block, e.g. ".foo {"
	selectorRegex := regexp.MustCompile(`^([^{}]+)\{`)

	var currentSelector string
	var insideBlock bool

	// Track whether we’re inside a media query
	var insideMedia bool
	var mediaQueryContext string
	var mediaBraceDepth int

	// Track whether we’re inside a keyframes block
	var insideKeyframes bool
	var keyframesBraceDepth int

	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		trimmed := strings.TrimSpace(line)

		// Check if we start an @media block
		if strings.HasPrefix(trimmed, "@media") {
			insideMedia = true
			mediaQueryContext = parseMediaQuery(trimmed)
			mediaBraceDepth = 0
			if strings.Contains(trimmed, "{") {
				mediaBraceDepth++
			}
			continue
		}
		// If inside a @media block, track braces
		if insideMedia {
			// Update brace depth for this line
			mediaBraceDepth += strings.Count(trimmed, "{")
			mediaBraceDepth -= strings.Count(trimmed, "}")
			if mediaBraceDepth <= 0 {
				// We've closed the @media block
				insideMedia = false
				mediaQueryContext = ""
			}
		}

		// Check if we start an @keyframes block
		// e.g. "@keyframes fadeIn {"
		// or "@-webkit-keyframes fadeIn {"
		if strings.HasPrefix(trimmed, "@keyframes") ||
			strings.HasPrefix(trimmed, "@-webkit-keyframes") ||
			strings.HasPrefix(trimmed, "@-moz-keyframes") ||
			strings.HasPrefix(trimmed, "@-o-keyframes") {
			insideKeyframes = true
			keyframesBraceDepth = 0
			if strings.Contains(trimmed, "{") {
				keyframesBraceDepth++
			}
			// We skip processing inside keyframes entirely,
			// or you can store them in a separate structure if you like.
			continue
		}
		// If inside keyframes, skip lines until we close the block
		if insideKeyframes {
			keyframesBraceDepth += strings.Count(trimmed, "{")
			keyframesBraceDepth -= strings.Count(trimmed, "}")
			if keyframesBraceDepth <= 0 {
				insideKeyframes = false
			}
			// Skip processing lines inside keyframes
			continue
		}

		// If we're here, we're not inside a keyframes block
		// Check if this line starts a new selector block
		if selectorRegex.MatchString(trimmed) {
			matches := selectorRegex.FindStringSubmatch(trimmed)
			if len(matches) > 1 {
				selectors := strings.Split(matches[1], ",")
				// For simplicity, track the first selector as 'current'
				currentSelector = strings.TrimSpace(selectors[0])
				insideBlock = true
			}
		} else if insideBlock {
			// We are inside a selector block
			// Check if we reached a closing brace
			if strings.Contains(trimmed, "}") {
				insideBlock = false
				currentSelector = ""
				continue
			}

			// Attempt to parse a property-value pair: "color: red;" or "background: #fff;"
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				prop := strings.TrimSpace(parts[0])
				val := parts[1]
				// remove trailing ';'
				val = strings.TrimSuffix(strings.TrimSpace(val), ";")

				storeCSSDeclaration(
					cssInfo,
					currentSelector,
					prop,
					val,
					filePath,
					lineNum,
					mediaQueryContext,
				)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error in file %s: %v\n", filePath, err)
	}
}

// parseMediaQuery tries to extract the entire `@media` line up to the first "{"
func parseMediaQuery(line string) string {
	idx := strings.Index(line, "{")
	if idx != -1 {
		return strings.TrimSpace(line[:idx])
	}
	return strings.TrimSpace(line)
}

// storeCSSDeclaration updates the global map with a newly found declaration
func storeCSSDeclaration(
	cssInfo CSSInfo,
	selector,
	property,
	value,
	filePath string,
	line int,
	mediaQuery string,
) {
	if _, ok := cssInfo[selector]; !ok {
		cssInfo[selector] = make(map[string][]ConflictDetail)
	}

	details := cssInfo[selector][property]
	details = append(details, ConflictDetail{
		FilePath:   filePath,
		Line:       line,
		Value:      value,
		MediaQuery: mediaQuery,
	})
	cssInfo[selector][property] = details
}

// collectConflicts goes through the entire CSSInfo and identifies
// if a given selector-property pair has multiple distinct values.
func collectConflicts(cssInfo CSSInfo) map[string]map[string][]ConflictDetail {
	// Return structure similar to CSSInfo, but only including conflicts
	conflicts := make(map[string]map[string][]ConflictDetail)

	for selector, propMap := range cssInfo {
		for prop, conflictDetails := range propMap {
			distinctValues := make(map[string][]ConflictDetail)
			for _, detail := range conflictDetails {
				distinctValues[detail.Value] = append(distinctValues[detail.Value], detail)
			}
			if len(distinctValues) > 1 {
				// We have a conflict because there's more than one distinct value
				if conflicts[selector] == nil {
					conflicts[selector] = make(map[string][]ConflictDetail)
				}
				var combined []ConflictDetail
				for _, detSlice := range distinctValues {
					combined = append(combined, detSlice...)
				}
				conflicts[selector][prop] = combined
			}
		}
	}

	return conflicts
}

// printConflicts prints out conflicts in a simple textual format to the given writer.
func printConflicts(conflicts map[string]map[string][]ConflictDetail, w io.Writer, folderDepth int) {
	if len(conflicts) == 0 {
		fmt.Fprintln(w, "No CSS conflicts found.")
		return
	}

	fmt.Fprintln(w, "=== CSS Conflicts Detected ===")
	for selector, propMap := range conflicts {
		fmt.Fprintf(w, "Selector: %s\n", selector)
		for prop, details := range propMap {
			fmt.Fprintf(w, "  Property: %s\n", prop)

			// group by value to show all distinct values and where they occur
			distinctByValue := make(map[string][]ConflictDetail)
			for _, d := range details {
				distinctByValue[d.Value] = append(distinctByValue[d.Value], d)
			}

			for val, valDetails := range distinctByValue {
				fmt.Fprintf(w, "    Value: %s\n", val)
				for _, vd := range valDetails {
					shortPath := shortenPath(vd.FilePath, folderDepth)
					if vd.MediaQuery != "" {
						// Indicate if we were inside a media query
						fmt.Fprintf(w, "      -> %s:%d (media: %s)\n", shortPath, vd.Line, vd.MediaQuery)
					} else {
						fmt.Fprintf(w, "      -> %s:%d\n", shortPath, vd.Line)
					}
				}
			}
		}
		fmt.Fprintln(w)
	}
}

// shortenPath keeps only the last `folderDepth` segments of `fullPath`.
func shortenPath(fullPath string, folderDepth int) string {
	parts := strings.Split(fullPath, "/")

	// Filter out empty parts if leading slash is present
	filtered := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	// If folderDepth >= length of filtered, return "/" + join(all)
	if folderDepth >= len(filtered) {
		return "/" + strings.Join(filtered, "/")
	}

	// Keep last folderDepth segments
	lastSegments := filtered[len(filtered)-folderDepth:]
	return "/" + strings.Join(lastSegments, "/")
}
