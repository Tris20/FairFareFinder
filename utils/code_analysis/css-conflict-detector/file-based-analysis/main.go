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

// CSSInfo holds all discovered CSS declarations
type CSSInfo map[string]map[string][]ConflictDetail

// ConflictDetail stores info about where/how a property was declared
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

	// Track if we're currently inside a multi-line comment
	var inComment bool

	lineNum := 0
	for scanner.Scan() {
		rawLine := scanner.Text()
		lineNum++

		// First strip out any comment text
		cleanLine, stillInComment := stripComments(rawLine, inComment)
		inComment = stillInComment

		// If after removing comments the line is blank or we remain inside a comment, skip
		cleanLine = strings.TrimSpace(cleanLine)
		if cleanLine == "" || inComment {
			continue
		}

		// We can now do the naive parse (media, keyframes, selectors, etc.) with `cleanLine`
		trimmed := cleanLine

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
			mediaBraceDepth += strings.Count(trimmed, "{")
			mediaBraceDepth -= strings.Count(trimmed, "}")
			if mediaBraceDepth <= 0 {
				insideMedia = false
				mediaQueryContext = ""
			}
		}

		// Check if we start an @keyframes block
		// e.g. "@keyframes fadeIn {"
		if strings.HasPrefix(trimmed, "@keyframes") ||
			strings.HasPrefix(trimmed, "@-webkit-keyframes") ||
			strings.HasPrefix(trimmed, "@-moz-keyframes") ||
			strings.HasPrefix(trimmed, "@-o-keyframes") {
			insideKeyframes = true
			keyframesBraceDepth = 0
			if strings.Contains(trimmed, "{") {
				keyframesBraceDepth++
			}
			// We skip processing inside keyframes entirely
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

		// If we're here, we’re not in keyframes block. Check if this line starts a new selector block.
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
			if strings.Contains(trimmed, "}") {
				insideBlock = false
				currentSelector = ""
				continue
			}

			// Attempt to parse a property-value pair: "color: red;" or "background: #fff;"
			parts := strings.Split(trimmed, ":")
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

// stripComments removes anything inside /* ... */ from a single line,
// allowing for partial inline removal. Also manages multi-line comment state.
func stripComments(line string, inComment bool) (string, bool) {
	var sb strings.Builder
	i := 0
	for i < len(line) {
		// If we're currently inside a comment, look for the closing '*/'
		if inComment {
			idx := strings.Index(line[i:], "*/")
			if idx == -1 {
				// No closing '*/' this line, skip entire remainder
				return "", true
			} else {
				// Found closing '*/'
				i += idx + 2 // move past the '*/'
				inComment = false
				continue
			}
		} else {
			// Not in a comment, look for '/*'
			idx := strings.Index(line[i:], "/*")
			if idx == -1 {
				// No comment start, append remainder to output
				sb.WriteString(line[i:])
				break
			} else {
				// Append everything up to the '/*'
				sb.WriteString(line[i : i+idx])
				// Move index to after '/*'
				i += idx + 2
				inComment = true
				// Now we loop again and look for '*/'
			}
		}
	}
	return sb.String(), inComment
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
	filtered := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	if folderDepth >= len(filtered) {
		return "/" + strings.Join(filtered, "/")
	}
	lastSegments := filtered[len(filtered)-folderDepth:]
	return "/" + strings.Join(lastSegments, "/")
}
