package kong

import (
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"
)

type Tag struct {
	Cmd         bool
	Arg         bool
	Required    bool
	Optional    bool
	Help        string
	Default     string
	Format      string
	Placeholder string
	Env         string
	Short       rune
}

func parseCSV(s string) ([]string, error) {
	num := 0
	parts := []string{}
	current := []rune{}

	next := func() {
		parts = append(parts, string(current))
		current = []rune{}
		num ++
	}

	quotes := false

	runes := []rune(s)
	for idx := 0; idx < len(runes); idx++ {
		r := runes[idx]
		next1 := uint8(0)
		eof := false
		if idx < len(s) - 1 {
			next1 = s[idx+1]
		} else {
			eof = true
		}
		if !quotes && r == ',' {
			next()
			continue
		}
		if r == '\\' {
			if next1 == '\'' {
				idx++
				r = '\''
			}
		} else if r == '\'' {
			if quotes {
				quotes = false
				if next1 == ',' || eof {
					continue
				}
				return parts, fmt.Errorf("%v has an unexpected char at pos %v", s, idx)
			} else {
				quotes = true
				continue
			}
		}
		current = append(current, r)
	}
	if quotes {
		return parts, fmt.Errorf("%v is not quoted properly", s)
	}

	next()

	return parts, nil
}

func parseTag(fv reflect.Value, s string) (*Tag, error) {
	t := &Tag{}
	if s == "" {
		return t, nil
	}

	parts, err := parseCSV(s)
	if err != nil {
		return t, err
	}

	for _, part := range parts {
		is := func(m string) bool { return part == m }
		value := func(m string) (string, bool) {
			split := strings.SplitN(part, "=", 2)
			if split[0] != m {
				return "", false
			}
			if len(split) == 1 {
				return "", true
			}
			return split[1], true
		}

		if is("cmd") {
			t.Cmd = true
		} else if is("arg") {
			t.Arg = true
		} else if is("required") {
			t.Required = true
		} else if is("optional") {
			t.Optional = true
		} else if v, ok := value("default"); ok {
			t.Default = v
		} else if v, ok := value("help"); ok {
			t.Help = v
		} else if v, ok := value("placeholder"); ok {
			t.Placeholder = v
		} else if v, ok := value("env"); ok {
			t.Env = v
		} else if v, ok := value("rune"); ok {
			t.Short, _ = utf8.DecodeRuneInString(v)
			if t.Short == utf8.RuneError {
				t.Short = 0
			}
		} else {
			return t, fmt.Errorf("%v is an unknown kong key", part)
		}
	}

	if t.Placeholder == "" {
		t.Placeholder = strings.ToUpper(dashedString(fv.Type().Name()))
	}

	return t, nil
}
