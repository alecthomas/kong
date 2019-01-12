package kong_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/alecthomas/kong"
)

// nolint: govet
type threeArg struct {
	RequiredThree bool   `required`
	Three         string `arg`
}

func (threeArg) Help() string {
	return `Detailed help provided through the HelpProvider interface.`
}

func TestHelp(t *testing.T) {
	// nolint: govet
	var cli struct {
		String   string         `help:"A string flag."`
		Bool     bool           `help:"A bool flag with very long help that wraps a lot and is verbose and is really verbose."`
		Slice    []string       `help:"A slice of strings." placeholder:"STR"`
		Map      map[string]int `help:"A map of strings to ints."`
		Required bool           `required help:"A required flag."`

		One struct {
			Flag string `help:"Nested flag."`
		} `cmd help:"A subcommand."`

		Two struct {
			Flag        string `help:"Nested flag under two."`
			RequiredTwo bool   `required`

			Three threeArg `arg help:"Sub-sub-arg."`

			Four struct {
			} `cmd help:"Sub-sub-command."`
		} `cmd help:"Another subcommand."`
	}

	w := bytes.NewBuffer(nil)
	exited := false
	app := mustNew(t, &cli,
		kong.Name("test-app"),
		kong.Description("A test app."),
		kong.Writers(w, w),
		kong.Exit(func(int) {
			exited = true
			panic(true) // Panic to fake "exit".
		}),
	)

	t.Run("Full", func(t *testing.T) {
		require.PanicsWithValue(t, true, func() {
			_, err := app.Parse([]string{"--help"})
			require.NoError(t, err)
		})
		require.True(t, exited)
		expected := `Usage: test-app --required <command>

A test app.

Flags:
  --help                 Show context-sensitive help.
  --string=STRING        A string flag.
  --bool                 A bool flag with very long help that wraps a lot and is
                         verbose and is really verbose.
  --slice=STR,...        A slice of strings.
  --map=KEY=VALUE;...    A map of strings to ints.
  --required             A required flag.

Commands:
  one --required
    A subcommand.

  two <three> --required --required-two --required-three
    Sub-sub-arg.

  two four --required --required-two
    Sub-sub-command.

Run "test-app <command> --help" for more information on a command.
`
		t.Log(w.String())
		t.Log(expected)
		require.Equal(t, expected, w.String())
	})

	t.Run("Selected", func(t *testing.T) {
		exited = false
		w.Truncate(0)
		require.PanicsWithValue(t, true, func() {
			_, err := app.Parse([]string{"two", "hello", "--help"})
			require.NoError(t, err)
		})
		require.True(t, exited)
		expected := `Usage: test-app two <three> --required --required-two --required-three

Sub-sub-arg.

Detailed help provided through the HelpProvider interface.

Flags:
  --help                 Show context-sensitive help.
  --string=STRING        A string flag.
  --bool                 A bool flag with very long help that wraps a lot and is
                         verbose and is really verbose.
  --slice=STR,...        A slice of strings.
  --map=KEY=VALUE;...    A map of strings to ints.
  --required             A required flag.

  --flag=STRING          Nested flag under two.
  --required-two

  --required-three
`
		t.Log(expected)
		t.Log(w.String())
		require.Equal(t, expected, w.String())
	})
}
