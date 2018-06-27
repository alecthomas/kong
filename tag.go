package kong

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Tag represents the parsed state of Kong tags in a struct field tag.
type Tag struct {
	Cmd         bool
	Arg         bool
	Required    bool
	Optional    bool
	Name        string
	Help        string
	Type        string
	Default     string
	Format      string
	PlaceHolder string
	Env         string
	Short       rune
	Hidden      bool
	Sep         rune
	Enum        string

	// Storage for all tag keys for arbitrary lookups.
	items map[string]string
}

type tagChars struct {
	sep, quote, assign rune
}

var kongChars = tagChars{sep: ',', quote: '\'', assign: '='}
var bareChars = tagChars{sep: ' ', quote: '"', assign: ':'}

func parseTagItems(s string, chr tagChars) map[string]string {
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
		if !quotes && r == chr.sep {
			add()
			continue
		}
		if r == chr.assign && inKey {
			inKey = false
			continue
		}
		if r == '\\' {
			if next == chr.quote {
				idx++
				r = chr.quote
			}
		} else if r == chr.quote {
			if quotes {
				quotes = false
				if next == chr.sep || eof {
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

func getTagInfo(ft reflect.StructField) (string, tagChars) {
	s, ok := ft.Tag.Lookup("kong")
	if ok {
		return s, kongChars
	}

	return string(ft.Tag), bareChars
}

func parseTag(fv reflect.Value, ft reflect.StructField) *Tag {
	s, chars := getTagInfo(ft)
	t := &Tag{
		items: parseTagItems(s, chars),
	}
	t.Cmd = t.Has("cmd")
	t.Arg = t.Has("arg")
	required := t.Has("required")
	optional := t.Has("optional")
	if required && optional {
		fail("can't specify both required and optional")
	}
	t.Required = required
	t.Optional = optional
	t.Default = t.Get("default")
	// Arguments with defaults are always optional.
	if t.Arg && t.Default != "" {
		t.Optional = true
	}
	t.Name = t.Get("name")
	t.Help = t.Get("help")
	t.Type = t.Get("type")
	t.Env = t.Get("env")
	t.Short, _ = t.GetRune("short")
	t.Hidden = t.Has("hidden")
	t.Format = t.Get("format")
	t.Sep, _ = t.GetRune("sep")
	if t.Sep == 0 {
		if t.Get("sep") == "none" {
			t.Sep = -1
		} else {
			t.Sep = ','
		}
	}
	t.PlaceHolder = t.Get("placeholder")
	if t.PlaceHolder == "" {
		t.PlaceHolder = strings.ToUpper(dashedString(fv.Type().Name()))
	}
	t.Enum = t.Get("enum")
	return t
}

// Has returns true if the tag contained the given key.
func (t *Tag) Has(k string) bool {
	_, ok := t.items[k]
	return ok
}

// Get returns the value of the given tag.
//
// Note that this will return the empty string if the tag is missing.
func (t *Tag) Get(k string) string {
	return t.items[k]
}

// GetBool returns true if the given tag looks like a boolean truth string.
func (t *Tag) GetBool(k string) (bool, error) {
	return strconv.ParseBool(t.items[k])
}

// GetFloat parses the given tag as a float64.
func (t *Tag) GetFloat(k string) (float64, error) {
	return strconv.ParseFloat(t.items[k], 64)
}

// GetInt parses the given tag as an int64.
func (t *Tag) GetInt(k string) (int64, error) {
	return strconv.ParseInt(t.items[k], 10, 64)
}

// GetRune parses the given tag as a rune.
func (t *Tag) GetRune(k string) (rune, error) {
	r, _ := utf8.DecodeRuneInString(t.items[k])
	if r == utf8.RuneError {
		return 0, fmt.Errorf("%v has a rune error", t.items[k])
	}
	return r, nil
}
