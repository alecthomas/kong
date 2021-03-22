package kong

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// ParseENVFile reads an env file from io.Reader and returns a kay/value map
func ParseENVFile(r io.Reader) (map[string]string, error) {
	envMap := make(map[string]string)

	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	err := scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("couldn't read an .env file: %w", err)
	}

	for _, fullLine := range lines {
		if !isIgnoredLine(fullLine) {
			var key, value string
			key, value, err = parseLine(fullLine, envMap)

			if err != nil {
				return nil, fmt.Errorf("couldn't parse a line of an .env file: %w", err)
			}
			envMap[key] = value
		}
	}
	return envMap, nil
}

func parseLine(line string, envMap map[string]string) (key string, value string, err error) {
	if len(line) == 0 {
		return "", "", errors.New("empty string")
	}

	line = uncommentLine(line)

	splitString := strings.SplitN(line, "=", 2)
	if len(splitString) != 2 {
		return "", "", errors.New("couldn't separate key from value")
	}

	// Parse the key
	key = splitString[0]
	// Get rid of spaces and "export …"
	key = strings.Trim(strings.TrimPrefix(key, "export"), " ")

	// Parse the value
	value = parseValue(splitString[1], envMap)
	return key, value, nil
}

func parseValue(value string, envMap map[string]string) string {
	value = strings.Trim(value, " ")
	if len(value) <= 1 {
		return value
	}

	// check if we've got quoted values or possible escapes
	regexpSingle := regexp.MustCompile(`\A'(.*)'\z`)
	singleQuotes := regexpSingle.FindStringSubmatch(value)

	regexpDouble := regexp.MustCompile(`\A"(.*)"\z`)
	doubleQuotes := regexpDouble.FindStringSubmatch(value)

	if singleQuotes != nil || doubleQuotes != nil {
		// Remove the quotes around the edge of the line
		value = value[1 : len(value)-1]
	}

	if doubleQuotes != nil {
		// expand newlines
		escapeRegex := regexp.MustCompile(`\\.`)
		value = escapeRegex.ReplaceAllStringFunc(value, func(match string) string {
			c := strings.TrimPrefix(match, `\`)
			switch c {
			case "n":
				return "\n"
			case "r":
				return "\r"
			default:
				return match
			}
		})
		// unescape characters
		e := regexp.MustCompile(`\\([^$])`)
		value = e.ReplaceAllString(value, "$1")
	}

	if singleQuotes == nil {
		value = expandVariables(value, envMap)
	}

	return value
}

func expandVariables(v string, m map[string]string) string {
	r := regexp.MustCompile(`(\\)?(\$)(\()?\{?([A-Z0-9_]+)?\}?`)

	return r.ReplaceAllStringFunc(v, func(s string) string {
		submatch := r.FindStringSubmatch(s)

		if submatch == nil {
			return s
		}
		if submatch[1] == "\\" || submatch[2] == "(" {
			return submatch[0][1:]
		} else if submatch[4] != "" {
			return m[submatch[4]]
		}
		return s
	})
}

// uncommentLine removes commented segments from line
func uncommentLine(line string) string {
	if !strings.Contains(line, "#") {
		return line
	}

	// Get rid of comment segments, but keep quoted # signs
	segmentsBetweenHashes := strings.Split(line, "#")
	quotesAreOpen := false
	var segmentsToKeep []string
	for _, segment := range segmentsBetweenHashes {
		if strings.Count(segment, "\"") == 1 || strings.Count(segment, "'") == 1 {
			if quotesAreOpen {
				quotesAreOpen = false
				segmentsToKeep = append(segmentsToKeep, segment)
			} else {
				quotesAreOpen = true
			}
		}

		if len(segmentsToKeep) == 0 || quotesAreOpen {
			segmentsToKeep = append(segmentsToKeep, segment)
		}
	}

	return strings.Join(segmentsToKeep, "#")
}

// isIgnoredLine checks whether line is a comment
func isIgnoredLine(line string) bool {
	trimmedLine := strings.Trim(line, " \n\t")
	return len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "#")
}
