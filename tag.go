package kong

import (
	"encoding/csv"
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

func trimQuotes(s string) string {
	if len(s) < 2 {
		return s
	}
	if s[0] == '\'' && s[len(s)-1] == '\'' {
		return s[1 : len(s)-1]
	}
	return s
}

func parseTag(fv reflect.Value, s string) (*Tag, error) {
	t := &Tag{}
	if s == "" {
		return t, nil
	}

	r := csv.NewReader(strings.NewReader(s))
	parts, err := r.Read()
	if err != nil {
		return t, fmt.Errorf("could not parse kong tag because %v", err)
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
			return trimQuotes(split[1]), true
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
