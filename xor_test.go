package kong_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/kong"
)

func TestXorPositional(t *testing.T) {
	tests := []struct {
		name string
		args []string
		err  string
	}{
		{name: "Neither"},
		{name: "Flag", args: []string{"--defaults"}},
		{name: "Positional", args: []string{"one", "two"}},
		{name: "FlagThenPositional", args: []string{"--defaults", "one"}, err: "--defaults and [<args> ...] can't be used together"},
		{name: "PositionalThenFlag", args: []string{"one", "--defaults"}, err: "--defaults and [<args> ...] can't be used together"},
		{name: "ExplicitFalse", args: []string{"--defaults=false", "one"}, err: "--defaults and [<args> ...] can't be used together"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var cli struct {
				Defaults bool     `xor:"source"`
				Args     []string `arg:"" optional:"" xor:"source"`
			}
			_, err := mustNew(t, &cli).Parse(test.args)
			if test.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.err)
			}
		})
	}
}

func TestXorPositionalRepeatedParse(t *testing.T) {
	for _, test := range []struct {
		name     string
		first    []string
		second   []string
		err      string
		defaults bool
		arg      string
	}{
		{name: "PositionalToFlag", first: []string{"one"}, second: []string{"--defaults"}, defaults: true},
		{name: "FlagToPositional", first: []string{"--defaults"}, second: []string{"one"}, arg: "one"},
		{name: "PositionalToNeither", first: []string{"one"}},
		{name: "FlagToNeither", first: []string{"--defaults"}},
		{name: "ConflictAfterPositional", first: []string{"one"}, second: []string{"--defaults", "two"}, err: "--defaults and [<arg>] can't be used together", defaults: true, arg: "two"},
		{name: "ConflictAfterFlag", first: []string{"--defaults"}, second: []string{"--defaults=false", ""}, err: "--defaults and [<arg>] can't be used together"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var cli struct {
				Defaults bool   `xor:"source"`
				Arg      string `arg:"" optional:"" xor:"source"`
			}
			p := mustNew(t, &cli)
			_, err := p.Parse(test.first)
			assert.NoError(t, err)
			_, err = p.Parse(test.second)
			if test.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.err)
			}
			assert.Equal(t, test.defaults, cli.Defaults)
			assert.Equal(t, test.arg, cli.Arg)
			_, err = p.Parse(nil)
			assert.NoError(t, err)
			assert.False(t, cli.Defaults)
			assert.Equal(t, "", cli.Arg)
		})
	}
}

func TestXorRequiredRepeatedParse(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
	}{
		{name: "Flag", args: []string{"--defaults"}},
		{name: "Positional", args: []string{"one"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Run("RequiredPositional", func(t *testing.T) {
				var cli struct {
					Defaults bool   `xor:"source"`
					Arg      string `arg:"" xor:"source"`
				}
				p := mustNew(t, &cli)
				_, err := p.Parse(test.args)
				assert.NoError(t, err)
				_, err = p.Parse(nil)
				assert.EqualError(t, err, `expected "<arg>"`)
			})
			t.Run("RequiredFlag", func(t *testing.T) {
				var cli struct {
					Defaults bool   `required:"" xor:"source"`
					Arg      string `arg:"" optional:"" xor:"source"`
				}
				p := mustNew(t, &cli)
				_, err := p.Parse(test.args)
				assert.NoError(t, err)
				_, err = p.Parse(nil)
				assert.EqualError(t, err, "missing flags or arguments: --defaults or [<arg>]")
			})
		})
	}
}

func TestXorPositionalRepeatedParseEnv(t *testing.T) {
	var cli struct {
		Defaults bool   `xor:"source"`
		Arg      string `arg:"" optional:"" xor:"source" env:"KONG_XOR_REPARSE_ARG"`
	}
	t.Setenv("KONG_XOR_REPARSE_ARG", "from-env")
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	assert.NoError(t, err)
	assert.Equal(t, "from-env", cli.Arg)
	_, err = p.Parse([]string{"--defaults"})
	assert.EqualError(t, err, "--defaults and [<arg>] can't be used together")
	assert.NoError(t, os.Unsetenv("KONG_XOR_REPARSE_ARG"))
	_, err = p.Parse([]string{"--defaults"})
	assert.NoError(t, err)
	assert.True(t, cli.Defaults)
	assert.Equal(t, "", cli.Arg)
}

func TestXorRequiredPositional(t *testing.T) {
	for _, args := range [][]string{{"--defaults"}, {"--defaults=false"}, {"one", "two"}, {""}} {
		var cli struct {
			Defaults bool     `xor:"source"`
			Args     []string `arg:"" xor:"source"`
		}
		_, err := mustNew(t, &cli).Parse(args)
		assert.NoError(t, err)
	}
	var cli struct {
		Defaults bool     `xor:"source"`
		Args     []string `arg:"" xor:"source"`
	}
	_, err := mustNew(t, &cli).Parse(nil)
	assert.EqualError(t, err, `expected "<args> ..."`)
}

func TestXorRequiredPositionalWithOptionalTail(t *testing.T) {
	var cli struct {
		Defaults bool   `xor:"source"`
		Arg      string `arg:"" xor:"source"`
		Tail     string `arg:"" optional:""`
	}
	_, err := mustNew(t, &cli).Parse([]string{"--defaults"})
	assert.NoError(t, err)
}

func TestXorRequiredFlagWithPositional(t *testing.T) {
	for _, args := range [][]string{{"--defaults"}, {"one", "two"}} {
		var cli struct {
			Defaults bool     `required:"" xor:"source"`
			Args     []string `arg:"" optional:"" xor:"source"`
		}
		_, err := mustNew(t, &cli).Parse(args)
		assert.NoError(t, err)
	}
	var cli struct {
		Defaults bool     `required:"" xor:"source"`
		Args     []string `arg:"" optional:"" xor:"source"`
	}
	_, err := mustNew(t, &cli).Parse(nil)
	assert.EqualError(t, err, "missing flags or arguments: --defaults or [<args> ...]")
}

func TestXorPositionals(t *testing.T) {
	var cli struct {
		First  string `arg:"" xor:"source"`
		Second string `arg:"" optional:"" xor:"source"`
	}
	_, err := mustNew(t, &cli).Parse([]string{"one", "two"})
	assert.EqualError(t, err, "<first> and [<second>] can't be used together")
}

func TestXorPositionalNodeScope(t *testing.T) {
	tests := []struct {
		name string
		args []string
		err  string
	}{
		{name: "ParentFlag", args: []string{"--global", "run", "one"}},
		{name: "InheritedFlag", args: []string{"run", "one", "--global"}},
		{name: "ParentAndChildFlags", args: []string{"--global", "run", "--local"}},
		{name: "ChildConflict", args: []string{"run", "--local", "one"}, err: "--local and [<args> ...] can't be used together"},
		{name: "Sibling", args: []string{"--global", "other", "one"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var cli struct {
				Global bool `xor:"source"`
				Run    struct {
					Local bool     `xor:"source"`
					Args  []string `arg:"" optional:"" xor:"source"`
				} `cmd:""`
				Other struct {
					Args []string `arg:"" xor:"source"`
				} `cmd:""`
			}
			_, err := mustNew(t, &cli).Parse(test.args)
			if test.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.err)
			}
		})
	}
}

func TestXorBranchingPositional(t *testing.T) {
	var cli struct {
		Global bool `xor:"source"`
		Source struct {
			Source string `arg:"" xor:"source"`
			Local  bool   `xor:"source"`
		} `arg:""`
	}
	_, err := mustNew(t, &cli).Parse([]string{"one", "--local"})
	assert.EqualError(t, err, "<source> and --local can't be used together")
	_, err = mustNew(t, &cli).Parse([]string{"--global", "one"})
	assert.NoError(t, err)
}

func TestXorPositionalMultipleRequiredGroups(t *testing.T) {
	for _, test := range []struct {
		args []string
		err  string
	}{
		{args: []string{"--one", "--two"}},
		{args: []string{"value"}},
		{args: []string{"--one"}, err: `expected "<arg>"`},
	} {
		var cli struct {
			One bool   `xor:"one"`
			Two bool   `xor:"two"`
			Arg string `arg:"" xor:"one,two"`
		}
		_, err := mustNew(t, &cli).Parse(test.args)
		if test.err == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, test.err)
		}
	}
}

func TestXorFlagSatisfiesMultipleRequiredPositionals(t *testing.T) {
	for _, test := range []struct {
		args []string
		err  string
	}{
		{args: []string{"--defaults"}},
		{args: []string{"one", "two"}},
		{args: []string{"one"}, err: `expected "<second>"`},
	} {
		var cli struct {
			Defaults bool   `xor:"one,two"`
			First    string `arg:"" xor:"one"`
			Second   string `arg:"" xor:"two"`
		}
		_, err := mustNew(t, &cli).Parse(test.args)
		if test.err == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, test.err)
		}
	}
}

func TestXorRequiredFlagMultipleGroupsWithPositional(t *testing.T) {
	var cli struct {
		One   bool   `required:"" xor:"one,two"`
		Two   bool   `required:"" xor:"one"`
		Three bool   `required:"" xor:"two"`
		Arg   string `arg:"" optional:"" xor:"one"`
	}
	_, err := mustNew(t, &cli).Parse([]string{"value"})
	assert.EqualError(t, err, "missing flags: --one or --three")
}

func TestXorRequiredPositionalNodeScope(t *testing.T) {
	var cli struct {
		Global bool `xor:"source"`
		Run    struct {
			Arg string `arg:"" xor:"source"`
		} `cmd:""`
	}
	_, err := mustNew(t, &cli).Parse([]string{"--global", "run"})
	assert.EqualError(t, err, `expected "<arg>"`)
}

func TestXorRequiredFlagNodeScope(t *testing.T) {
	var cli struct {
		One bool `required:"" xor:"source"`
		Two bool `required:"" xor:"source"`
		Run struct {
			Arg string `arg:"" xor:"source"`
		} `cmd:""`
	}
	_, err := mustNew(t, &cli).Parse([]string{"run", "value"})
	assert.EqualError(t, err, "missing flags: --one or --two")
}

func TestXorPositionalPrefix(t *testing.T) {
	type sourceOptions struct {
		Defaults bool   `xor:"source"`
		Arg      string `arg:"" optional:"" xor:"source"`
	}
	var cli struct {
		Source sourceOptions `embed:"" prefix:"source-" xorprefix:"source-"`
		Other  bool          `xor:"source"`
	}
	_, err := mustNew(t, &cli).Parse([]string{"--source-defaults", "one"})
	assert.EqualError(t, err, "--source-defaults and [<source-arg>] can't be used together")
	_, err = mustNew(t, &cli).Parse([]string{"--other", "one"})
	assert.NoError(t, err)
}

func TestXorPreservesPositionalBinding(t *testing.T) {
	var cli struct {
		Skip   bool   `xor:"first"`
		First  string `arg:"" xor:"first"`
		Second string `arg:""`
	}
	_, err := mustNew(t, &cli).Parse([]string{"--skip", "value"})
	assert.EqualError(t, err, `expected "<second>"`)
	assert.Equal(t, "value", cli.First)
	assert.Equal(t, "", cli.Second)
}

func TestXorAndDistinctValuesWithSameName(t *testing.T) {
	var cli struct {
		Arg    string `arg:"" optional:"" name:"same" xor:"source"`
		Flag   bool   `name:"same" and:"pair"`
		Shared bool   `xor:"source" and:"pair"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--same", "--shared"})
	assert.NoError(t, err)
}

func TestXorAndChildModelCompatibility(t *testing.T) {
	var cli struct {
		Run struct {
			One bool `xor:"source" and:"pair"`
			Two bool `xor:"source" and:"pair"`
		} `cmd:""`
		Other struct{} `cmd:""`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"other"})
	assert.NoError(t, err)
}

func TestXorPositionalDefault(t *testing.T) {
	var cli struct {
		Defaults bool   `xor:"source"`
		Arg      string `arg:"" default:"fallback" xor:"source"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--defaults"})
	assert.EqualError(t, err, "--defaults and [<arg>] can't be used together")
	_, err = p.Parse(nil)
	assert.NoError(t, err)
	assert.Equal(t, "fallback", cli.Arg)
}

func TestXorFlagDefaultWithPositional(t *testing.T) {
	var cli struct {
		Defaults bool   `default:"true" xor:"source"`
		Arg      string `arg:"" optional:"" xor:"source"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	assert.NoError(t, err)
	_, err = p.Parse([]string{"one"})
	assert.EqualError(t, err, "--defaults and [<arg>] can't be used together")
	_, err = p.Parse(nil)
	assert.NoError(t, err)
	assert.True(t, cli.Defaults)
}

func TestXorPositionalUsesFlagModel(t *testing.T) {
	var cli struct {
		Defaults bool   `xor:"source"`
		Arg      string `arg:"" optional:"" xor:"source"`
	}
	p := mustNew(t, &cli)
	for _, flag := range p.Model.Flags {
		if flag.Name == "defaults" {
			flag.Xor = nil
		}
	}
	_, err := p.Parse([]string{"--defaults", "one"})
	assert.NoError(t, err)
}

func TestXorPositionalHelp(t *testing.T) {
	var cli struct {
		Defaults bool     `xor:"source" help:"Use defaults."`
		Args     []string `arg:"" xor:"source" help:"Source arguments."`
	}
	w := &bytes.Buffer{}
	p := mustNew(t, &cli, kong.Name("test"), kong.Writers(w, w), kong.Exit(func(int) { panic(true) }))
	panicsTrue(t, func() {
		_, _ = p.Parse([]string{"--help"})
	})
	assert.Equal(t, `Usage: test <args> ... [flags]

Arguments:
  <args> ...    Source arguments.

Flags:
  -h, --help        Show context-sensitive help.
      --defaults    Use defaults.
`, w.String())
}
