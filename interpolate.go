package kong

import "regexp"

var interpolationRegex = regexp.MustCompile(`(\$\$)|((?:\${([[:alpha:]_][[:word:]]*))(?:=([^}]+))?})|(\$)|([^$]+)`)

// HasInterpolatedVar returns true if the variable "v" is interpolated in "s".
func HasInterpolatedVar(s string, v string) bool {
	matches := interpolationRegex.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		if name := match[3]; name == v {
			return true
		}
	}
	return false
}

// Interpolate variables from vars into s for substrings in the form ${var} or ${var=default}.
func interpolate(s string, vars Vars, updatedVars map[string]string) (out string, err error) {
	var (
		matches  [][]string
		match    []string
		key, val string
		dollar   string
		name     string
		value    string
		ok       bool
	)

	if matches = interpolationRegex.FindAllStringSubmatch(s, -1); len(matches) == 0 {
		out = s
		return
	}
	for key, val = range updatedVars {
		if vars[key] != val {
			vars = vars.CloneWith(updatedVars)
			break
		}
	}
	for _, match = range matches {
		if dollar = match[1]; dollar != "" {
			out += delimiterDollar
		} else if name = match[3]; name != "" {
			if value, ok = vars[name]; !ok {
				// No default value.
				if match[4] == "" {
					err = Errors().UndefinedVariable(name)
					return
				}
				value = match[4]
			}
			out += value
		} else {
			out += match[0]
		}
	}

	return
}
