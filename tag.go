package kong

import (
	"fmt"
	"reflect"
	"strconv"
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

	// Storage for all tag keys for arbitrary lookups.
	items map[string]string
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
	t := &Tag{
		items: map[string]string{},
	}
	if s == "" {
		return t
	}

	t.items = parseCSV(s)

	t.Cmd = t.Has("cmd")
	t.Arg = t.Has("arg")
	t.Required = t.Has("required")
	t.Optional = t.Has("optional")
	t.Default, _ = t.Get("default")
	t.Help, _ = t.Get("help")
	t.Type, _ = t.Get("type")
	t.Env, _ = t.Get("env")
	t.Short, _ = t.GetRune("short")

	t.Placeholder, _ = t.Get("placeholder")
	if t.Placeholder == "" {
		t.Placeholder = strings.ToUpper(dashedString(fv.Type().Name()))
	}

	return t
}

func (t *Tag) Has(k string) bool {
	_, ok := t.items[k]
	return ok
}

func (t *Tag) Get(k string) (string, bool) {
	s, ok := t.items[k]
	return s, ok
}

func (t *Tag) GetBool(k string) (bool, error) {
	return strconv.ParseBool(t.items[k])
}

func (t *Tag) GetFloat(k string) (float64, error) {
	return strconv.ParseFloat(t.items[k], 64)
}

func (t *Tag) GetInt(k string) (int64, error) {
	return strconv.ParseInt(t.items[k], 10, 64)
}

func (t *Tag) GetRune(k string) (rune, error) {
	r, _ := utf8.DecodeRuneInString(t.items[k])
	if r == utf8.RuneError {
		return 0, fmt.Errorf("%v has a rune error", t.items[k])
	}
	return r, nil
}
