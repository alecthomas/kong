package kong_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/kong"
)

func panicsTrue(t *testing.T, f func()) {
	defer func() {
		if value := recover(); value != nil {
			if boolval, ok := value.(bool); !ok || !boolval {
				t.Fatalf("expected panic with true but got %v", value)
			}
		}
	}()
	f()
	t.Fatal("expected panic did not occur")
}

type threeArg struct {
	RequiredThree bool   `required`
	Three         string `arg`
}

func (threeArg) Help() string {
	return `Detailed help provided through the HelpProvider interface.`
}

func TestHelpOptionalArgs(t *testing.T) {
	var cli struct {
		One string `arg:"" optional:"" help:"One optional arg."`
		Two string `arg:"" optional:"" help:"Two optional arg."`
	}
	w := bytes.NewBuffer(nil)
	exited := false
	app := mustNew(t, &cli,
		kong.Name("test-app"),
		kong.Writers(w, w),
		kong.Exit(func(int) {
			exited = true
			panic(true) // Panic to fake "exit".
		}),
	)
	panicsTrue(t, func() {
		_, err := app.Parse([]string{"--help"})
		assert.NoError(t, err)
	})
	assert.True(t, exited)
	expected := `Usage: test-app [<one> [<two>]] [flags]

Arguments:
  [<one>]    One optional arg.
  [<two>]    Two optional arg.

Flags:
  -h, --help    Show context-sensitive help.
`
	assert.Equal(t, expected, w.String())
}

func TestHelp(t *testing.T) {
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
		panicsTrue(t, func() {
			_, err := app.Parse([]string{"--help"})
			assert.NoError(t, err)
		})
		assert.True(t, exited)
		expected := `Usage: test-app --required <command> [flags]

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
  one --required [flags]
    A subcommand.

  two <three> --required --required-two --required-three [flags]
    Sub-sub-arg.

  two four --required --required-two [flags]
    Sub-sub-command.

Run "test-app <command> --help" for more information on a command.
`
		t.Log(w.String())
		t.Log(expected)
		assert.Equal(t, expected, w.String())
	})

	t.Run("Selected", func(t *testing.T) {
		exited = false
		w.Truncate(0)
		panicsTrue(t, func() {
			_, err := app.Parse([]string{"two", "hello", "--help"})
			assert.NoError(t, err)
		})
		assert.True(t, exited)
		expected := `Usage: test-app two <three> --required --required-two --required-three [flags]

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
		assert.Equal(t, expected, w.String())
	})
}

func TestFlagsLast(t *testing.T) {
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
		panicsTrue(t, func() {
			_, err := app.Parse([]string{"--help"})
			assert.NoError(t, err)
		})
		assert.True(t, exited)
		expected := `Usage: test-app --required <command> [flags]

A test app.

Commands:
  one --required [flags]
    A subcommand.

  two <three> --required --required-two --required-three [flags]
    Sub-sub-arg.

  two four --required --required-two [flags]
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
		assert.Equal(t, expected, w.String())
	})

	t.Run("Selected", func(t *testing.T) {
		exited = false
		w.Truncate(0)
		panicsTrue(t, func() {
			_, err := app.Parse([]string{"two", "hello", "--help"})
			assert.NoError(t, err)
		})
		assert.True(t, exited)
		expected := `Usage: test-app two <three> --required --required-two --required-three [flags]

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
		assert.Equal(t, expected, w.String())
	})
}

func TestHelpTree(t *testing.T) {
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
		panicsTrue(t, func() {
			_, err := app.Parse([]string{"--help"})
			assert.NoError(t, err)
		})
		assert.True(t, exited)
		expected := `Usage: test-app <command> [flags]

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
		assert.Equal(t, expected, w.String())
	})

	t.Run("Selected", func(t *testing.T) {
		exited = false
		w.Truncate(0)
		panicsTrue(t, func() {
			_, err := app.Parse([]string{"one", "--help"})
			assert.NoError(t, err)
		})
		assert.True(t, exited)
		expected := `Usage: test-app one (un,uno) <command> [flags]

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
		assert.Equal(t, expected, w.String())
	})
}

func TestHelpCompactNoExpand(t *testing.T) {
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
		panicsTrue(t, func() {
			_, err := app.Parse([]string{"--help"})
			assert.NoError(t, err)
		})
		assert.True(t, exited)
		expected := `Usage: test-app <command> [flags]

A test app.

Flags:
  -h, --help    Show context-sensitive help.

Commands:
  two    Another subcommand.

Group A
  one (un,uno)    subcommand one

Run "test-app <command> --help" for more information on a command.
`
		if expected != w.String() {
			t.Errorf("help command returned:\n%v\n\nwant:\n%v", w.String(), expected)
		}
		assert.Equal(t, expected, w.String())
	})

	t.Run("Selected", func(t *testing.T) {
		exited = false
		w.Truncate(0)
		panicsTrue(t, func() {
			_, err := app.Parse([]string{"one", "--help"})
			assert.NoError(t, err)
		})
		assert.True(t, exited)
		expected := `Usage: test-app one (un,uno) <command> [flags]

subcommand one

Flags:
  -h, --help    Show context-sensitive help.

Group A
  one (un,uno) thing      subcommand thing
  one (un,uno) <other>    subcommand other
`
		if expected != w.String() {
			t.Errorf("help command returned:\n%v\n\nwant:\n%v", w.String(), expected)
		}
		assert.Equal(t, expected, w.String())
	})
}

func TestEnvarAutoHelp(t *testing.T) {
	var cli struct {
		Flag string `env:"FLAG" help:"A flag."`
	}
	w := &strings.Builder{}
	p := mustNew(t, &cli, kong.Writers(w, w), kong.Exit(func(int) {}))
	_, err := p.Parse([]string{"--help"})
	assert.NoError(t, err)
	assert.Contains(t, w.String(), "A flag ($FLAG).")
}

func TestMultipleEnvarAutoHelp(t *testing.T) {
	var cli struct {
		Flag string `env:"FLAG1,FLAG2" help:"A flag."`
	}
	w := &strings.Builder{}
	p := mustNew(t, &cli, kong.Writers(w, w), kong.Exit(func(int) {}))
	_, err := p.Parse([]string{"--help"})
	assert.NoError(t, err)
	assert.Contains(t, w.String(), "A flag ($FLAG1, $FLAG2).")
}

//nolint:dupl // false positive
func TestEnvarAutoHelpWithEnvPrefix(t *testing.T) {
	type Anonymous struct {
		Flag  string `env:"FLAG" help:"A flag."`
		Other string `help:"A different flag."`
	}
	var cli struct {
		Anonymous `envprefix:"ANON_"`
	}
	w := &strings.Builder{}
	p := mustNew(t, &cli, kong.Writers(w, w), kong.Exit(func(int) {}))
	_, err := p.Parse([]string{"--help"})
	assert.NoError(t, err)
	assert.Contains(t, w.String(), "A flag ($ANON_FLAG).")
	assert.Contains(t, w.String(), "A different flag.")
}

//nolint:dupl // false positive
func TestMultipleEnvarAutoHelpWithEnvPrefix(t *testing.T) {
	type Anonymous struct {
		Flag  string `env:"FLAG1,FLAG2" help:"A flag."`
		Other string `help:"A different flag."`
	}
	var cli struct {
		Anonymous `envprefix:"ANON_"`
	}
	w := &strings.Builder{}
	p := mustNew(t, &cli, kong.Writers(w, w), kong.Exit(func(int) {}))
	_, err := p.Parse([]string{"--help"})
	assert.NoError(t, err)
	assert.Contains(t, w.String(), "A flag ($ANON_FLAG1, $ANON_FLAG2).")
	assert.Contains(t, w.String(), "A different flag.")
}

//nolint:dupl // false positive
func TestCustomValueFormatter(t *testing.T) {
	var cli struct {
		Flag string `env:"FLAG" help:"A flag."`
	}
	w := &strings.Builder{}
	p := mustNew(t, &cli,
		kong.Writers(w, w),
		kong.Exit(func(int) {}),
		kong.ValueFormatter(func(value *kong.Value) string {
			return value.Help
		}),
	)
	_, err := p.Parse([]string{"--help"})
	assert.NoError(t, err)
	assert.Contains(t, w.String(), "A flag.")
}

//nolint:dupl // false positive
func TestMultipleCustomValueFormatter(t *testing.T) {
	var cli struct {
		Flag string `env:"FLAG1,FLAG2" help:"A flag."`
	}
	w := &strings.Builder{}
	p := mustNew(t, &cli,
		kong.Writers(w, w),
		kong.Exit(func(int) {}),
		kong.ValueFormatter(func(value *kong.Value) string {
			return value.Help
		}),
	)
	_, err := p.Parse([]string{"--help"})
	assert.NoError(t, err)
	assert.Contains(t, w.String(), "A flag.")
}

func TestAutoGroup(t *testing.T) {
	var cli struct {
		GroupedAString string `help:"A string flag grouped in A."`
		FreeString     string `help:"A non grouped string flag."`
		GroupedBString string `help:"A string flag grouped in B."`
		FreeBool       bool   `help:"A non grouped bool flag."`
		GroupedABool   bool   `help:"A bool flag grouped in A."`

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
			} `arg help:"subcommand stuff"`
		} `cmd help:"A subcommand grouped in A."`

		Two struct {
			Grouped1String  string `help:"A string flag grouped in 1."`
			AFreeString     string `help:"A non grouped string flag."`
			Grouped2String  string `help:"A string flag grouped in 2."`
			AGroupedAString bool   `help:"A string flag grouped in A."`
			Grouped1Bool    bool   `help:"A bool flag grouped in 1."`
		} `cmd help:"A non grouped subcommand."`

		Four struct {
			Flag string `help:"Nested flag."`
		} `cmd help:"Another subcommand grouped in B."`

		Three struct {
			Flag string `help:"Nested flag."`
		} `cmd help:"Another subcommand grouped in A."`
	}
	w := bytes.NewBuffer(nil)
	app := mustNew(t, &cli,
		kong.Writers(w, w),
		kong.Exit(func(int) {}),
		kong.AutoGroup(func(parent kong.Visitable, flag *kong.Flag) *kong.Group {
			if node, ok := parent.(*kong.Node); ok {
				return &kong.Group{
					Key:   node.Name,
					Title: strings.Title(node.Name) + " flags:", //nolint
				}
			}
			return nil
		}),
	)
	_, _ = app.Parse([]string{"--help", "two"})
	assert.Equal(t, `Usage: test two [flags]

A non grouped subcommand.

Flags:
  -h, --help                       Show context-sensitive help.
      --grouped-a-string=STRING    A string flag grouped in A.
      --free-string=STRING         A non grouped string flag.
      --grouped-b-string=STRING    A string flag grouped in B.
      --free-bool                  A non grouped bool flag.
      --grouped-a-bool             A bool flag grouped in A.

Two flags:
  --grouped-1-string=STRING    A string flag grouped in 1.
  --a-free-string=STRING       A non grouped string flag.
  --grouped-2-string=STRING    A string flag grouped in 2.
  --a-grouped-a-string         A string flag grouped in A.
  --grouped-1-bool             A bool flag grouped in 1.
`, w.String())
}

func TestHelpGrouping(t *testing.T) {
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
		panicsTrue(t, func() {
			_, err := app.Parse([]string{"--help"})
			assert.True(t, exited)
			assert.NoError(t, err)
		})
		expected := `Usage: test-app <command> [flags]

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
  two [flags]
    A non grouped subcommand.

Group title taken from the kong.ExplicitGroups option
  A group header

  one thing <arg> [flags]
    subcommand thing

  one <other> [flags]
    subcommand other

  three [flags]
    Another subcommand grouped in A.

Group B
  one <stuff> [flags]
    subcommand stuff

  four [flags]
    Another subcommand grouped in B.

Run "test-app <command> --help" for more information on a command.
`
		t.Log(w.String())
		t.Log(expected)
		assert.Equal(t, expected, w.String())
	})

	t.Run("Selected", func(t *testing.T) {
		exited = false
		w.Truncate(0)
		panicsTrue(t, func() {
			_, err := app.Parse([]string{"two", "--help"})
			assert.NoError(t, err)
			assert.True(t, exited)
		})
		expected := `Usage: test-app two [flags]

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
		assert.Equal(t, expected, w.String())
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

	expected := `Usage: test --flag=STRING [flags]

Some description.

Flags:
  -h, --help           Show context-sensitive help.
      --flag=STRING    A required flag.

test: error: missing flags: --flag=STRING
`
	assert.Equal(t, expected, w.String())
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
	assert.Error(t, err)
	p.FatalIfErrorf(err)

	expected := `Usage: test --flag=STRING [flags]
Run "test --help" for more information.

test: error: missing flags: --flag=STRING
`
	assert.Equal(t, expected, w.String())
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
	assert.Error(t, err)
	p.FatalIfErrorf(err)

	expected := `🤷 wish I could help

test: error: missing flags: --flag=STRING
`
	assert.Equal(t, expected, w.String())
}
