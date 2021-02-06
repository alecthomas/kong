package kong_test

import (
	"bytes"
	"strings"
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
  -h, --help                 Show context-sensitive help.
      --string=STRING        A string flag.
      --bool                 A bool flag with very long help that wraps a lot
                             and is verbose and is really verbose.
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
  -h, --help                 Show context-sensitive help.
      --string=STRING        A string flag.
      --bool                 A bool flag with very long help that wraps a lot
                             and is verbose and is really verbose.
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

func TestHelpTree(t *testing.T) {
	// nolint: govet
	var cli struct {
		One struct {
			Thing struct {
				Arg string `arg help:"argument"`
			} `cmd help:"subcommand thing"`
			Other struct {
				Other string `arg help:"other arg"`
			} `arg help:"subcommand other"`
		} `cmd help:"subcommand one" group:"Group A"` // Groups are ignored in trees

		Two struct {
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
		kong.ConfigureHelp(kong.HelpOptions{
			Tree:     true,
			Indenter: kong.LineIndenter,
		}),
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
		expected := `Usage: test-app <command>

A test app.

Flags:
  -h, --help    Show context-sensitive help.

Commands:
  one          subcommand one
  - thing      subcommand thing
    - <arg>    argument
  - <other>    subcommand other

  two          Another subcommand.
  - <three>    Sub-sub-arg.
  - four       Sub-sub-command.

Run "test-app <command> --help" for more information on a command.
`
		if expected != w.String() {
			t.Errorf("help command returned:\n%v\n\nwant:\n%v", w.String(), expected)
		}
		require.Equal(t, expected, w.String())
	})

	t.Run("Selected", func(t *testing.T) {
		exited = false
		w.Truncate(0)
		require.PanicsWithValue(t, true, func() {
			_, err := app.Parse([]string{"one", "--help"})
			require.NoError(t, err)
		})
		require.True(t, exited)
		expected := `Usage: test-app one <command>

subcommand one

Flags:
  -h, --help    Show context-sensitive help.

Commands:
  thing      subcommand thing
  - <arg>    argument

  <other>    subcommand other
`
		if expected != w.String() {
			t.Errorf("help command returned:\n%v\n\nwant:\n%v", w.String(), expected)
		}
		require.Equal(t, expected, w.String())
	})
}

func TestEnvarAutoHelp(t *testing.T) {
	var cli struct {
		Flag string `env:"FLAG" help:"A flag."`
	}
	w := &strings.Builder{}
	p := mustNew(t, &cli, kong.Writers(w, w), kong.Exit(func(int) {}))
	_, err := p.Parse([]string{"--help"})
	require.NoError(t, err)
	require.Contains(t, w.String(), "A flag ($FLAG).")
}

func TestCustomHelpFormatter(t *testing.T) {
	var cli struct {
		Flag string `env:"FLAG" help:"A flag."`
	}
	w := &strings.Builder{}
	p := mustNew(t, &cli,
		kong.Writers(w, w),
		kong.Exit(func(int) {}),
		kong.HelpFormatter(func(value *kong.Value) string {
			return value.Help
		}),
	)
	_, err := p.Parse([]string{"--help"})
	require.NoError(t, err)
	require.Contains(t, w.String(), "A flag.")
}

func TestHelpGrouping(t *testing.T) {
	// nolint: govet
	var cli struct {
		GroupedAString string `help:"A string flag grouped in A." group:"Group A"`
		FreeString     string `help:"A non grouped string flag."`
		GroupedBString string `help:"A string flag grouped in B." group:"Group B"`
		FreeBool       bool   `help:"A non grouped bool flag."`
		GroupedABool   bool   `help:"A bool flag grouped in A." group:"Group A"`

		One struct {
			Flag string `help:"Nested flag."`
			// Group is inherited from the parent command
			Thing struct {
				Arg string `arg help:"argument"`
			} `cmd help:"subcommand thing"`
			Other struct {
				Other string `arg help:"other arg"`
			} `arg help:"subcommand other"`
			// ... but a subcommand can override it
			Stuff struct {
				Stuff string `arg help:"argument"`
			} `arg help:"subcommand stuff" group:"Group B"`
		} `cmd help:"A subcommand grouped in A." group:"Group A"`

		Two struct {
			Grouped1String string `help:"A string flag grouped in A." group:"Group 1"`
			AFreeString    string `help:"A non grouped string flag."`
			Grouped2String string `help:"A string flag grouped in B." group:"Group 2"`
			AFreeBool      bool   `help:"A non grouped bool flag."`
			Grouped1Bool   bool   `help:"A bool flag grouped in A." group:"Group 1"`
		} `cmd help:"A non grouped subcommand."`

		Four struct {
			Flag string `help:"Nested flag."`
		} `cmd help:"Another subcommand grouped in B." group:"Group B"`

		Three struct {
			Flag string `help:"Nested flag."`
		} `cmd help:"Another subcommand grouped in A." group:"Group A"`
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
			require.True(t, exited)
			require.NoError(t, err)
		})
		expected := `Usage: test-app <command>

A test app.

Flags:
  -h, --help                       Show context-sensitive help.
      --grouped-a-string=STRING    A string flag grouped in A.
      --free-string=STRING         A non grouped string flag.
      --grouped-b-string=STRING    A string flag grouped in B.
      --free-bool                  A non grouped bool flag.
      --grouped-a-bool             A bool flag grouped in A.

Commands:
  two
    A non grouped subcommand.

Group A:
  one thing <arg>
    subcommand thing

  one <other>
    subcommand other

  three
    Another subcommand grouped in A.

Group B:
  one <stuff>
    subcommand stuff

  four
    Another subcommand grouped in B.

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
			_, err := app.Parse([]string{"two", "--help"})
			require.NoError(t, err)
			require.True(t, exited)
		})
		expected := `Usage: test-app two

A non grouped subcommand.

Flags:
  -h, --help                       Show context-sensitive help.
      --grouped-a-string=STRING    A string flag grouped in A.
      --free-string=STRING         A non grouped string flag.
      --grouped-b-string=STRING    A string flag grouped in B.
      --free-bool                  A non grouped bool flag.
      --grouped-a-bool             A bool flag grouped in A.

      --grouped-1-string=STRING    A string flag grouped in A.
      --a-free-string=STRING       A non grouped string flag.
      --grouped-2-string=STRING    A string flag grouped in B.
      --a-free-bool                A non grouped bool flag.
      --grouped-1-bool             A bool flag grouped in A.
`
		t.Log(expected)
		t.Log(w.String())
		require.Equal(t, expected, w.String())
	})
}
