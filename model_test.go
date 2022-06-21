package kong_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
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
	assert.Equal(t, []string{"one two", "one three <four>"}, actual)
}

func TestFlagString(t *testing.T) {
	var cli struct {
		String                  string
		DefaultInt              int    `default:"42"`
		DefaultStr              string `default:"hello"`
		Placeholder             string `placeholder:"world"`
		DefaultPlaceholder      string `default:"hello" placeholder:"world"`
		SliceSep                []string
		SliceNoSep              []string `sep:"none"`
		SliceDefault            []string `default:"hello"`
		SlicePlaceholder        []string `placeholder:"world"`
		SliceDefaultPlaceholder []string `default:"hello" placeholder:"world"`
		MapSep                  map[string]string
		MapNoSep                map[string]string `mapsep:"none"`
		MapDefault              map[string]string `mapsep:"none" default:"hello"`
		MapPlaceholder          map[string]string `mapsep:"none" placeholder:"world"`
		Counter                 int               `type:"counter"`
	}
	tests := map[string]string{
		"help":                      "-h, --help",
		"string":                    "--string=STRING",
		"default-int":               "--default-int=42",
		"default-str":               `--default-str="hello"`,
		"placeholder":               "--placeholder=world",
		"default-placeholder":       "--default-placeholder=world",
		"slice-sep":                 "--slice-sep=SLICE-SEP,...",
		"slice-no-sep":              "--slice-no-sep=SLICE-NO-SEP",
		"slice-default":             "--slice-default=hello,...",
		"slice-placeholder":         "--slice-placeholder=world,...",
		"slice-default-placeholder": "--slice-default-placeholder=world,...",
		"map-sep":                   "--map-sep=KEY=VALUE;...",
		"map-no-sep":                "--map-no-sep=KEY=VALUE",
		"map-default":               "--map-default=hello",
		"map-placeholder":           "--map-placeholder=world",
		"counter":                   "--counter",
	}

	p := mustNew(t, &cli)
	for _, flag := range p.Model.Flags {
		want, ok := tests[flag.Name]
		assert.True(t, ok, "unknown flag name: %s", flag.Name)
		assert.Equal(t, want, flag.String())
	}
}
