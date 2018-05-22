package kong

import (
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
	Type        string
	Default     string
	Format      string
	Placeholder string
	Env         string
	Short       rune
}

func parseCSV(s string) map[string]string {
	d := map[string]string{}

	key := []rune{}
	value := []rune{}
	quotes := false
	inKey := true

	add := func() {
		d[string(key)] = string(value)
		key = []rune{}
		value = []rune{}
		inKey = true
	}

	runes := []rune(s)
	for idx := 0; idx < len(runes); idx++ {
		r := runes[idx]
		next := rune(0)
		eof := false
		if idx < len(runes)-1 {
			next = runes[idx+1]
		} else {
			eof = true
		}
		if !quotes && r == ',' {
			add()
			continue
		}
		if r == '=' && inKey {
			inKey = false
			continue
		}
		if r == '\\' {
			if next == '\'' {
				idx++
				r = '\''
			}
		} else if r == '\'' {
			if quotes {
				quotes = false
				if next == ',' || eof {
					continue
				}
				fail("%v has an unexpected char at pos %v", s, idx)
			} else {
				quotes = true
				continue
			}
		}
		if inKey {
			key = append(key, r)
		} else {
			value = append(value, r)
		}
	}
	if quotes {
		fail("%v is not quoted properly", s)
	}

	add()

	return d
}

func parseTag(fv reflect.Value, s string) *Tag {
	t := &Tag{}
	if s == "" {
		return t
	}

	for k, v := range parseCSV(s) {
		switch k {
		case "cmd":
			t.Cmd = true
		case "arg":
			t.Arg = true
		case "required":
			t.Required = true
		case "optional":
			t.Optional = true
		case "default":
			t.Default = v
		case "help":
			t.Help = v
		case "type":
			t.Type = v
		case "placeholder":
			t.Placeholder = v
		case "env":
			t.Env = v
		case "rune":
			t.Short, _ = utf8.DecodeRuneInString(v)
			if t.Short == utf8.RuneError {
				t.Short = 0
			}
		default:
			fail("%v is an unknown kong key", k)
		}
	}

	if t.Placeholder == "" {
		t.Placeholder = strings.ToUpper(dashedString(fv.Type().Name()))
	}

	return t
}
