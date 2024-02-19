
package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ConvertMarkdownToHTML converts markdown content to HTML and saves it to the specified file path.
func ConvertMarkdownToHTML(markdownContent, outputPath string) error {
	// Ensure the output directory exists
	if err := os.MkdirAll(strings.TrimSuffix(outputPath, "/berlin-flight-destinations.html"), os.ModePerm); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	// Create or open the output file for writing
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer outputFile.Close()

	// Use bufio.NewWriter for efficient writing
	writer := bufio.NewWriter(outputFile)

	// Start writing the HTML content
	writer.WriteString("<table>\n")
	scanner := bufio.NewScanner(strings.NewReader(markdownContent))
	isHeader := true
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "|") {
			line = strings.Trim(line, "|")
			columns := strings.Split(line, "|")

			if isHeader {
				writer.WriteString("  <tr>\n")
				for _, col := range columns {
					writer.WriteString(fmt.Sprintf("    <th>%s</th>\n", strings.TrimSpace(col)))
				}
				writer.WriteString("  </tr>\n")
				isHeader = false
			} else {
				writer.WriteString("  <tr>\n")
				for _, col := range columns {
					col = strings.TrimSpace(col)
					col = markdownLinkToHTML(col)
					writer.WriteString(fmt.Sprintf("    <td>%s</td>\n", col))
				}
				writer.WriteString("  </tr>\n")
			}
		}
	}
	writer.WriteString("</table>\n")

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading markdown content: %w", err)
	}

	// Ensure all buffered content is written to the file
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing writer: %w", err)
	}

	return nil
}

// markdownLinkToHTML converts markdown links to HTML links.
func markdownLinkToHTML(text string) string {
	re := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	return re.ReplaceAllString(text, `<a href="$2">$1</a>`)
}
