package htmltablegenerator

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func ConvertMarkdownToHTML(markdownContent, outputPath string) error {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)

	// Initial HTML structure
	writer.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Markdown Table to HTML</title>
    <link rel="stylesheet" href="../tableStyles.css">
</head>
<body>
<table>
`)

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
				for i, col := range columns {
					writer.WriteString(fmt.Sprintf("    <td>%s</td>\n", processColumnContent(strings.TrimSpace(col), i)))
				}
				writer.WriteString("  </tr>\n")
			}
		}
	}
	writer.WriteString("</table>\n</body>\n</html>")

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading markdown content: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing writer: %w", err)
	}

	return nil
}

// processColumnContent identifies image-link pairs and formats them as HTML.
func processColumnContent(text string, columnIndex int) string {
	if columnIndex == 1 { // Targeting the second column
		// This regular expression matches your unique Markdown-like syntax for images and links
		re := regexp.MustCompile(`\[\((.*?)\)\]\((.*?)\)`)
		return re.ReplaceAllString(text, `<a href="$2"><img src="$1" alt="Image" style="max-width:100px;"></a>`) // Adjust styling as necessary
	} else {
		// Handle other columns that may contain standard Markdown links
		re := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
		text = re.ReplaceAllString(text, `<a href="$2">$1</a>`)
	}
	return text
}
