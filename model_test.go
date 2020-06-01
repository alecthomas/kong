package kong_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModelApplicationCommands(t *testing.T) {
	var cli struct {
		One struct {
			Two struct {
			} `kong:"cmd"`
			Three struct {
				Four struct {
					Four string `kong:"arg"`
				} `kong:"arg"`
			} `kong:"cmd"`
		} `kong:"cmd"`
	}
	p := mustNew(t, &cli)
	actual := []string{}
	for _, cmd := range p.Model.Leaves(false) {
		actual = append(actual, cmd.Path())
	}
	require.Equal(t, []string{"one two", "one three <four>"}, actual)
}

func TestFormatPlaceHolder(t *testing.T) {
	var cli struct {
		String                  string
		DefaultInt              int    `default:"42"`
		DefaultStr              string `default:"hello"`
		Placeholder             string `placeholder:"world"`
		SliceSep                []string
		SliceNoSep              []string `sep:"none"`
		SliceDefault            []string `default:"hello"`
		SlicePlaceholder        []string `placeholder:"world"`
		SliceDefaultPlaceholder []string `default:"hello" placeholder:"world"`
		MapSep                  map[string]string
		MapNoSep                map[string]string `mapsep:"none"`
		MapDefault              map[string]string `mapsep:"none" default:"hello"`
		MapPlaceholder          map[string]string `mapsep:"none" placeholder:"world"`
	}
	tests := map[string]string{
		"help":                      "HELP",
		"string":                    "STRING",
		"default-int":               "42",
		"default-str":               `"hello"`,
		"placeholder":               "world",
		"slice-sep":                 "SLICE-SEP,...",
		"slice-no-sep":              "SLICE-NO-SEP",
		"slice-default":             "hello,...",
		"slice-placeholder":         "world,...",
		"slice-default-placeholder": "hello,...",
		"map-sep":                   "KEY=VALUE;...",
		"map-no-sep":                "KEY=VALUE",
		"map-default":               "hello",
		"map-placeholder":           "world",
	}

	p := mustNew(t, &cli)
	for _, flag := range p.Model.Flags {
		want, ok := tests[flag.Name]
		require.Truef(t, ok, "unknown flag name: %s", flag.Name)
		require.Equal(t, want, flag.FormatPlaceHolder())
	}
}
