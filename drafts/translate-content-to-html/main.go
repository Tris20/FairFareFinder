
package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	// Assuming the markdown content is stored in a file named "markdown.txt"
	file, err := os.Open("markdown.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Create or open the index.html file for writing
	outputFile, err := os.Create("index.html")
	if err != nil {
		fmt.Println("Error creating index.html:", err)
		return
	}
	defer outputFile.Close()

	// Use bufio.NewWriter for efficient writing
	writer := bufio.NewWriter(outputFile)

	// Writing the opening table tag to the file
	writer.WriteString("<table>\n")
	scanner := bufio.NewScanner(file)
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
					// Process each cell to convert markdown links to HTML links
					col = strings.TrimSpace(col)
					col = markdownLinkToHTML(col)
					writer.WriteString(fmt.Sprintf("    <td>%s</td>\n", col))
				}
				writer.WriteString("  </tr>\n")
			}
		}
	}
	// Writing the closing table tag to the file
	writer.WriteString("</table>\n")

	// Check for errors during scanning the file
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}

	// Important: Flush writes any buffered data to the underlying io.Writer
	writer.Flush()
}

// markdownLinkToHTML converts markdown links to HTML links.
func markdownLinkToHTML(text string) string {
	// Regular expression to match markdown links
	re := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	return re.ReplaceAllString(text, `<a href="$2">$1</a>`)
}

