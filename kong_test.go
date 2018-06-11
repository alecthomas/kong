package kong

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustNew(t *testing.T, cli interface{}, options ...Option) *Kong {
	t.Helper()
	options = append([]Option{
		ExitFunction(func(int) {
			t.Helper()
			t.Fatalf("unexpected exit()")
		}),
	}, options...)
	parser, err := New(cli, options...)
	require.NoError(t, err)
	return parser
}

func TestPositionalArguments(t *testing.T) {
	var cli struct {
		User struct {
			Create struct {
				ID    int    `kong:"arg"`
				First string `kong:"arg"`
				Last  string `kong:"arg"`
			} `kong:"cmd"`
		} `kong:"cmd"`
	}
	p := mustNew(t, &cli)
	cmd, err := p.Parse([]string{"user", "create", "10", "Alec", "Thomas"})
	require.NoError(t, err)
	require.Equal(t, "user create <id> <first> <last>", cmd)
	t.Run("Missing", func(t *testing.T) {
		_, err := p.Parse([]string{"user", "create", "10"})
		require.Error(t, err)
	})
}

func TestBranchingArgument(t *testing.T) {
	/*
		app user create <id> <first> <last>
		app	user <id> delete
		app	user <id> rename <to>

	*/
	var cli struct {
		User struct {
			Create struct {
				ID    string `kong:"arg"`
				First string `kong:"arg"`
				Last  string `kong:"arg"`
			} `kong:"cmd"`

			// Branching argument.
			ID struct {
				ID     int `kong:"arg"`
				Flag   int
				Delete struct{} `kong:"cmd"`
				Rename struct {
					To string
				} `kong:"cmd"`
			} `kong:"arg"`
		} `kong:"cmd,help='User management.'"`
	}
	p := mustNew(t, &cli)
	cmd, err := p.Parse([]string{"user", "10", "delete"})
	require.NoError(t, err)
	require.Equal(t, 10, cli.User.ID.ID)
	require.Equal(t, "user <id> delete", cmd)
	t.Run("Missing", func(t *testing.T) {
		_, err = p.Parse([]string{"user"})
		require.Error(t, err)
	})
}

func TestResetWithDefaults(t *testing.T) {
	var cli struct {
		Flag            string
		FlagWithDefault string `kong:"default='default'"`
	}
	cli.Flag = "BLAH"
	cli.FlagWithDefault = "BLAH"
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "", cli.Flag)
	require.Equal(t, "default", cli.FlagWithDefault)
}

func TestFlagSlice(t *testing.T) {
	var cli struct {
		Slice []int
	}
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{"--slice=1,2,3"})
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, cli.Slice)
}

func TestFlagSliceWithSeparator(t *testing.T) {
	var cli struct {
		Slice []string
	}
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{`--slice=a\,b,c`})
	require.NoError(t, err)
	require.Equal(t, []string{"a,b", "c"}, cli.Slice)
}

func TestArgSlice(t *testing.T) {
	var cli struct {
		Slice []int `arg`
		Flag  bool
	}
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{"1", "2", "3", "--flag"})
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, cli.Slice)
	require.Equal(t, true, cli.Flag)
}

func TestArgSliceWithSeparator(t *testing.T) {
	var cli struct {
		Slice []string `arg`
		Flag  bool
	}
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{"a,b", "c", "--flag"})
	require.NoError(t, err)
	require.Equal(t, []string{"a,b", "c"}, cli.Slice)
	require.Equal(t, true, cli.Flag)
}

func TestUnsupportedFieldErrors(t *testing.T) {
	var cli struct {
		Keys map[string]string
	}
	_, err := New(&cli)
	require.Error(t, err)
}

func TestMatchingArgField(t *testing.T) {
	var cli struct {
		ID struct {
			NotID int `kong:"arg"`
		} `kong:"arg"`
	}

	_, err := New(&cli)
	require.Error(t, err)
}

func TestCantMixPositionalAndBranches(t *testing.T) {
	var cli struct {
		Arg     string `kong:"arg"`
		Command struct {
		} `kong:"cmd"`
	}
	_, err := New(&cli)
	require.Error(t, err)
}

func TestPropagatedFlags(t *testing.T) {
	var cli struct {
		Flag1    string
		Command1 struct {
			Flag2    bool
			Command2 struct{} `kong:"cmd"`
		} `kong:"cmd"`
	}

	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{"command-1", "command-2", "--flag-2", "--flag-1=moo"})
	require.NoError(t, err)
	require.Equal(t, "moo", cli.Flag1)
	require.Equal(t, true, cli.Command1.Flag2)
}

func TestRequiredFlag(t *testing.T) {
	var cli struct {
		Flag string `kong:"required"`
	}

	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{})
	require.Error(t, err)
}

func TestOptionalArg(t *testing.T) {
	var cli struct {
		Arg string `kong:"arg,optional"`
	}

	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{})
	require.NoError(t, err)
}

func TestRequiredArg(t *testing.T) {
	var cli struct {
		Arg string `kong:"arg"`
	}

	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{})
	require.Error(t, err)
}

func TestInvalidRequiredAfterOptional(t *testing.T) {
	var cli struct {
		ID   int    `kong:"arg,optional"`
		Name string `kong:"arg"`
	}

	_, err := New(&cli)
	require.Error(t, err)
}

func TestOptionalStructArg(t *testing.T) {
	var cli struct {
		Name struct {
			Name    string `kong:"arg,optional"`
			Enabled bool
		} `kong:"arg,optional"`
	}

	parser := mustNew(t, &cli)

	t.Run("WithFlag", func(t *testing.T) {
		_, err := parser.Parse([]string{"gak", "--enabled"})
		require.NoError(t, err)
		require.Equal(t, "gak", cli.Name.Name)
		require.Equal(t, true, cli.Name.Enabled)
	})

	t.Run("WithoutFlag", func(t *testing.T) {
		_, err := parser.Parse([]string{"gak"})
		require.NoError(t, err)
		require.Equal(t, "gak", cli.Name.Name)
	})

	t.Run("WithNothing", func(t *testing.T) {
		_, err := parser.Parse([]string{})
		require.NoError(t, err)
	})
}

func TestMixedRequiredArgs(t *testing.T) {
	var cli struct {
		Name string `kong:"arg"`
		ID   int    `kong:"arg,optional"`
	}

	parser := mustNew(t, &cli)

	t.Run("SingleRequired", func(t *testing.T) {
		_, err := parser.Parse([]string{"gak", "5"})
		require.NoError(t, err)
		require.Equal(t, "gak", cli.Name)
		require.Equal(t, 5, cli.ID)
	})

	t.Run("ExtraOptional", func(t *testing.T) {
		_, err := parser.Parse([]string{"gak"})
		require.NoError(t, err)
		require.Equal(t, "gak", cli.Name)
	})
}

func TestInvalidDefaultErrors(t *testing.T) {
	var cli struct {
		Flag int `kong:"default='foo'"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	require.Error(t, err)
}

func TestCommandMissingTagIsInvalid(t *testing.T) {
	var cli struct {
		One struct{}
	}
	_, err := New(&cli)
	require.Error(t, err)
}

func TestDuplicateFlag(t *testing.T) {
	var cli struct {
		Flag bool
		Cmd  struct {
			Flag bool
		} `kong:"cmd"`
	}
	_, err := New(&cli)
	require.Error(t, err)
}

func TestDuplicateFlagOnPeerCommandIsOkay(t *testing.T) {
	var cli struct {
		Cmd1 struct {
			Flag bool
		} `kong:"cmd"`
		Cmd2 struct {
			Flag bool
		} `kong:"cmd"`
	}
	_, err := New(&cli)
	require.NoError(t, err)
}

func TestTraceErrorPartiallySucceeds(t *testing.T) {
	var cli struct {
		One struct {
			Two struct {
			} `kong:"cmd"`
		} `kong:"cmd"`
	}
	p := mustNew(t, &cli)
	ctx, err := p.Trace([]string{"one", "bad"})
	require.NoError(t, err)
	require.Error(t, ctx.Error)
	require.Equal(t, []string{"one"}, ctx.Command())
}

func TestHooks(t *testing.T) {
	var cli struct {
		One struct {
			Two   string `kong:"arg,optional"`
			Three string
		} `kong:"cmd"`
	}
	type values struct {
		one   bool
		two   string
		three string
	}
	hooked := values{}
	var tests = []struct {
		name   string
		input  string
		values values
	}{
		{"Command", "one", values{true, "", ""}},
		{"Arg", "one two", values{true, "two", ""}},
		{"Flag", "one --three=three", values{true, "", "three"}},
		{"ArgAndFlag", "one two --three=three", values{true, "two", "three"}},
	}
	setOne := func(ctx *Context, path *Path) error { hooked.one = true; return nil }
	setTwo := func(ctx *Context, path *Path) error { hooked.two = path.Value.String(); return nil }
	setThree := func(ctx *Context, path *Path) error { hooked.three = path.Value.String(); return nil }
	p := mustNew(t, &cli,
		Hook(&cli.One, setOne),
		Hook(&cli.One.Two, setTwo),
		Hook(&cli.One.Three, setThree))

	for _, test := range tests {
		hooked = values{}
		t.Run(test.name, func(t *testing.T) {
			_, err := p.Parse(strings.Split(test.input, " "))
			require.NoError(t, err)
			require.Equal(t, test.values, hooked)
		})
	}
}

func TestShort(t *testing.T) {
	var cli struct {
		Bool   bool   `short:"b"`
		String string `short:"s"`
	}
	app := mustNew(t, &cli)
	_, err := app.Parse([]string{"-b", "-shello"})
	require.NoError(t, err)
	require.True(t, cli.Bool)
	require.Equal(t, "hello", cli.String)
}

func TestDuplicateFlagChoosesLast(t *testing.T) {
	var cli struct {
		Flag int
	}

	_, err := mustNew(t, &cli).Parse([]string{"--flag=1", "--flag=2"})
	require.NoError(t, err)
	require.Equal(t, 2, cli.Flag)
}

func TestDuplicateSliceDoesNotAccumulate(t *testing.T) {
	var cli struct {
		Flag []int
	}

	_, err := mustNew(t, &cli).Parse([]string{"--flag=1,2", "--flag=3,4"})
	require.NoError(t, err)
	require.Equal(t, []int{3, 4}, cli.Flag)
}
