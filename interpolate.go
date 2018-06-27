package kong

import (
	"fmt"
	"regexp"
)

var interpolationRegex = regexp.MustCompile(`(\${[[:alpha:]_][[:word:]]*})|(\$)|([^$]+)`)

// Interpolate variables from vars into s for substrings in the form ${var}.
func interpolate(s string, vars map[string]string) (string, error) {
	out := ""
	matches := interpolationRegex.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		if match[1] != "" {
			name := match[1][2 : len(match[1])-1]
			value, ok := vars[name]
			if !ok {
				return "", fmt.Errorf("undefined variable ${%s}", name)
			}
			out += value
		} else {
			out += match[0]
		}
	}
	return out, nil
}
