package backend

import (
	"fmt"

	"errors"
	"strings"
	"time"
)

// Replace query placeholders (?) with arguments for debugging purposes.
func ReplacePlaceholdersWithArgs(query string, args []interface{}) (string, error) {
	var result strings.Builder
	argIndex := 0
	placeholderCount := strings.Count(query, "?")

	// Check for mismatch between placeholders and args
	if placeholderCount != len(args) {
		return "", errors.New("mismatch between placeholders and arguments")
	}

	for _, char := range query {
		if char == '?' && argIndex < len(args) {
			arg := args[argIndex]
			argIndex++

			switch v := arg.(type) {
			case string:
				// Escape single quotes to avoid breaking SQL syntax
				safeString := strings.ReplaceAll(v, "'", "''")
				result.WriteString(fmt.Sprintf("'%s'", safeString))
			case float64:
				result.WriteString(fmt.Sprintf("%.2f", v))
			case int:
				result.WriteString(fmt.Sprintf("%d", v))
			case nil:
				result.WriteString("NULL")
			case time.Time:
				// Format time as standard SQL datetime
				result.WriteString(fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05")))
			default:
				result.WriteString(fmt.Sprintf("%v", v))
			}
		} else {
			result.WriteRune(char)
		}
	}

	return result.String(), nil
}
