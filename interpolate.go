package kong

import (
	"fmt"
	"regexp"
)

var interpolationRegex = regexp.MustCompile(`((?:\${([[:alpha:]_][[:word:]]*))(?:=([^}]+))?})|(\$)|([^$]+)`)

// Returns true if the variable "v" is interpolated in "s".
func interpolationHasVar(s string, v string) bool {
	matches := interpolationRegex.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		if name := match[2]; name == v {
			return true
		}
	}
	return false
}

// Interpolate variables from vars into s for substrings in the form ${var} or ${var=default}.
func interpolate(s string, vars map[string]string) (string, error) {
	out := ""
	matches := interpolationRegex.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		if name := match[2]; name != "" {
			value, ok := vars[name]
			if !ok {
				// No default value.
				if match[3] == "" {
					return "", fmt.Errorf("undefined variable ${%s}", name)
				}
				value = match[3]
			}
			out += value
		} else {
			out += match[0]
		}
	}
	return out, nil
}
