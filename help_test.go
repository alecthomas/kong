package kong_test

import (
	"bytes"
	"fmt"
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
		Sort     bool           `negatable short:"s" help:"Is sortable or not."`

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
  -s, --[no-]sort            Is sortable or not.

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
  -s, --[no-]sort            Is sortable or not.

      --flag=STRING          Nested flag under two.
      --required-two

      --required-three
`
		t.Log(expected)
		t.Log(w.String())
		require.Equal(t, expected, w.String())
	})
}

func TestFlagsLast(t *testing.T) {
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
		kong.HelpOptions{
			FlagsLast: true,
		},
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

Commands:
  one --required
    A subcommand.

  two <three> --required --required-two --required-three
    Sub-sub-arg.

  two four --required --required-two
    Sub-sub-command.

Flags:
  -h, --help                 Show context-sensitive help.
      --string=STRING        A string flag.
      --bool                 A bool flag with very long help that wraps a lot
                             and is verbose and is really verbose.
      --slice=STR,...        A slice of strings.
      --map=KEY=VALUE;...    A map of strings to ints.
      --required             A required flag.

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
		} `cmd help:"subcommand one" group:"Group A" aliases:"un,uno"` // Groups are ignored in trees

		Two struct {
			Three threeArg `arg help:"Sub-sub-arg."`

			Four struct {
			} `cmd help:"Sub-sub-command." aliases:"for,fore"`
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
  one (un,uno)         subcommand one
  - thing              subcommand thing
    - <arg>            argument
  - <other>            subcommand other

  two                  Another subcommand.
  - <three>            Sub-sub-arg.
  - four (for,fore)    Sub-sub-command.

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

func TestHelpCompactNoExpand(t *testing.T) {
	// nolint: govet
	var cli struct {
		One struct {
			Thing struct {
				Arg string `arg help:"argument"`
			} `cmd help:"subcommand thing"`
			Other struct {
				Other string `arg help:"other arg"`
			} `arg help:"subcommand other"`
		} `cmd help:"subcommand one" group:"Group A" aliases:"un,uno"` // Groups are ignored in trees

		Two struct {
			Three threeArg `arg help:"Sub-sub-arg."`

			Four struct {
			} `cmd help:"Sub-sub-command." aliases:"for,fore"`
		} `cmd help:"Another subcommand."`
	}

	w := bytes.NewBuffer(nil)
	exited := false
	app := mustNew(t, &cli,
		kong.Name("test-app"),
		kong.Description("A test app."),
		kong.Writers(w, w),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:             true,
			NoExpandSubcommands: true,
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
  two    Another subcommand.

Group A
  one    subcommand one

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

Group A
  one thing      subcommand thing
  one <other>    subcommand other
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
			Grouped1String  string `help:"A string flag grouped in 1." group:"Group 1"`
			AFreeString     string `help:"A non grouped string flag."`
			Grouped2String  string `help:"A string flag grouped in 2." group:"Group 2"`
			AGroupedAString bool   `help:"A string flag grouped in A." group:"Group A"`
			Grouped1Bool    bool   `help:"A bool flag grouped in 1." group:"Group 1"`
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
		kong.Groups{
			"Group A":     "Group title taken from the kong.ExplicitGroups option\nA group header",
			"Group 1":     "Another group title, this time without header",
			"Unknown key": "",
		},
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
  -h, --help                  Show context-sensitive help.
      --free-string=STRING    A non grouped string flag.
      --free-bool             A non grouped bool flag.

Group title taken from the kong.ExplicitGroups option
  A group header

  --grouped-a-string=STRING    A string flag grouped in A.
  --grouped-a-bool             A bool flag grouped in A.

Group B
  --grouped-b-string=STRING    A string flag grouped in B.

Commands:
  two
    A non grouped subcommand.

Group title taken from the kong.ExplicitGroups option
  A group header

  one thing <arg>
    subcommand thing

  one <other>
    subcommand other

  three
    Another subcommand grouped in A.

Group B
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
  -h, --help                    Show context-sensitive help.
      --free-string=STRING      A non grouped string flag.
      --free-bool               A non grouped bool flag.

      --a-free-string=STRING    A non grouped string flag.

Group title taken from the kong.ExplicitGroups option
  A group header

  --grouped-a-string=STRING    A string flag grouped in A.
  --grouped-a-bool             A bool flag grouped in A.

  --a-grouped-a-string         A string flag grouped in A.

Group B
  --grouped-b-string=STRING    A string flag grouped in B.

Another group title, this time without header
  --grouped-1-string=STRING    A string flag grouped in 1.
  --grouped-1-bool             A bool flag grouped in 1.

Group 2
  --grouped-2-string=STRING    A string flag grouped in 2.
`
		t.Log(expected)
		t.Log(w.String())
		require.Equal(t, expected, w.String())
	})
}

func TestUsageOnError(t *testing.T) {
	var cli struct {
		Flag string `help:"A required flag." required`
	}
	w := &strings.Builder{}
	p := mustNew(t, &cli,
		kong.Writers(w, w),
		kong.Description("Some description."),
		kong.Exit(func(int) {}),
		kong.UsageOnError(),
	)
	_, err := p.Parse([]string{})
	p.FatalIfErrorf(err)

	expected := `Usage: test --flag=STRING

Some description.

Flags:
  -h, --help           Show context-sensitive help.
      --flag=STRING    A required flag.

test: error: missing flags: --flag=STRING
`
	require.Equal(t, expected, w.String())
}

func TestShortUsageOnError(t *testing.T) {
	var cli struct {
		Flag string `help:"A required flag." required`
	}
	w := &strings.Builder{}
	p := mustNew(t, &cli,
		kong.Writers(w, w),
		kong.Description("Some description."),
		kong.Exit(func(int) {}),
		kong.ShortUsageOnError(),
	)
	_, err := p.Parse([]string{})
	require.Error(t, err)
	p.FatalIfErrorf(err)

	expected := `Usage: test --flag=STRING
Run "test --help" for more information.

test: error: missing flags: --flag=STRING
`
	require.Equal(t, expected, w.String())
}

func TestCustomShortUsageOnError(t *testing.T) {
	var cli struct {
		Flag string `help:"A required flag." required`
	}
	w := &strings.Builder{}
	shortHelp := func(_ kong.HelpOptions, ctx *kong.Context) error {
		fmt.Fprintln(ctx.Stdout, "🤷 wish I could help")
		return nil
	}
	p := mustNew(t, &cli,
		kong.Writers(w, w),
		kong.Description("Some description."),
		kong.Exit(func(int) {}),
		kong.ShortHelp(shortHelp),
		kong.ShortUsageOnError(),
	)
	_, err := p.Parse([]string{})
	require.Error(t, err)
	p.FatalIfErrorf(err)

	expected := `🤷 wish I could help

test: error: missing flags: --flag=STRING
`
	require.Equal(t, expected, w.String())
}
