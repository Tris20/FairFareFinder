package backend

import (
	"fmt"
	"strings"
)

// Interpolates query placeholders (?) with arguments for debugging purposes.
func InterpolateQuery(query string, args []interface{}) string {
	var result strings.Builder
	argIndex := 0

	for _, char := range query {
		if char == '?' && argIndex < len(args) {
			arg := args[argIndex]
			argIndex++

			switch v := arg.(type) {
			case string:
				result.WriteString(fmt.Sprintf("'%s'", v))
			case float64:
				result.WriteString(fmt.Sprintf("%.2f", v))
			case int:
				result.WriteString(fmt.Sprintf("%d", v))
			default:
				result.WriteString(fmt.Sprintf("%v", v))
			}
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}
