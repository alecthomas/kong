package kong_test

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/alecthomas/kong"
)

func mustNew(t *testing.T, cli interface{}, options ...kong.Option) *kong.Kong {
	t.Helper()
	options = append([]kong.Option{
		kong.Name("test"),
		kong.Exit(func(int) {
			t.Helper()
			t.Fatalf("unexpected exit()")
		}),
	}, options...)
	parser, err := kong.New(cli, options...)
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
	ctx, err := p.Parse([]string{"user", "create", "10", "Alec", "Thomas"})
	require.NoError(t, err)
	require.Equal(t, "user create <id> <first> <last>", ctx.Command())
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
	ctx, err := p.Parse([]string{"user", "10", "delete"})
	require.NoError(t, err)
	require.Equal(t, 10, cli.User.ID.ID)
	require.Equal(t, "user <id> delete", ctx.Command())
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
		Keys struct{}
	}
	_, err := kong.New(&cli)
	require.Error(t, err)
}

func TestMatchingArgField(t *testing.T) {
	var cli struct {
		ID struct {
			NotID int `kong:"arg"`
		} `kong:"arg"`
	}

	_, err := kong.New(&cli)
	require.Error(t, err)
}

func TestCantMixPositionalAndBranches(t *testing.T) {
	var cli struct {
		Arg     string `kong:"arg"`
		Command struct {
		} `kong:"cmd"`
	}
	_, err := kong.New(&cli)
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

func TestOptionalArgWithDefault(t *testing.T) {
	var cli struct {
		Arg string `kong:"arg,optional,default='moo'"`
	}

	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "moo", cli.Arg)
}

func TestArgWithDefaultIsOptional(t *testing.T) {
	var cli struct {
		Arg string `kong:"arg,default='moo'"`
	}

	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "moo", cli.Arg)
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

	_, err := kong.New(&cli)
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
	_, err := kong.New(&cli)
	require.Error(t, err)
}

func TestDuplicateFlag(t *testing.T) {
	var cli struct {
		Flag bool
		Cmd  struct {
			Flag bool
		} `kong:"cmd"`
	}
	_, err := kong.New(&cli)
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
	_, err := kong.New(&cli)
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
	ctx, err := kong.Trace(p, []string{"one", "bad"})
	require.NoError(t, err)
	require.Error(t, ctx.Error)
	require.Equal(t, "one", ctx.Command())
}

type commandWithNegatableFlag struct {
	Flag bool `kong:"default='true',negatable"`
	ran  bool
}

func (c *commandWithNegatableFlag) Run() error {
	c.ran = true
	return nil
}

func TestNegatableFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "no flag",
			args:     []string{"cmd"},
			expected: true,
		},
		{
			name:     "boolean flag",
			args:     []string{"cmd", "--flag"},
			expected: true,
		},
		{
			name:     "inverted boolean flag",
			args:     []string{"cmd", "--flag=false"},
			expected: false,
		},
		{
			name:     "negated boolean flag",
			args:     []string{"cmd", "--no-flag"},
			expected: false,
		},
		{
			name:     "inverted negated boolean flag",
			args:     []string{"cmd", "--no-flag=false"},
			expected: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var cli struct {
				Cmd commandWithNegatableFlag `kong:"cmd"`
			}

			p := mustNew(t, &cli)
			kctx, err := p.Parse(tt.args)
			require.NoError(t, err)
			require.Equal(t, tt.expected, cli.Cmd.Flag)

			err = kctx.Run()
			require.NoError(t, err)
			require.Equal(t, tt.expected, cli.Cmd.Flag)
			require.True(t, cli.Cmd.ran)
		})
	}
}

func TestExistingNoFlag(t *testing.T) {
	var cli struct {
		Cmd struct {
			Flag   bool `kong:"default='true'"`
			NoFlag string
		} `kong:"cmd"`
	}

	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"cmd", "--no-flag=none"})
	require.NoError(t, err)
	require.Equal(t, true, cli.Cmd.Flag)
	require.Equal(t, "none", cli.Cmd.NoFlag)
}

func TestInvalidNegatedNonBool(t *testing.T) {
	var cli struct {
		Cmd struct {
			Flag string `kong:"negatable"`
		} `kong:"cmd"`
	}

	_, err := kong.New(&cli)
	require.Error(t, err)
}

type hookContext struct {
	cmd    bool
	values []string
}

type hookValue string

func (h *hookValue) BeforeApply(ctx *hookContext) error {
	ctx.values = append(ctx.values, "before:"+string(*h))
	return nil
}

func (h *hookValue) AfterApply(ctx *hookContext) error {
	ctx.values = append(ctx.values, "after:"+string(*h))
	return nil
}

type hookCmd struct {
	Two   hookValue `kong:"arg,optional"`
	Three hookValue
}

func (h *hookCmd) BeforeApply(ctx *hookContext) error {
	ctx.cmd = true
	return nil
}

func (h *hookCmd) AfterApply(ctx *hookContext) error {
	ctx.cmd = true
	return nil
}

func TestHooks(t *testing.T) {
	var tests = []struct {
		name   string
		input  string
		values hookContext
	}{
		{"Command", "one", hookContext{true, nil}},
		{"Arg", "one two", hookContext{true, []string{"before:", "after:two"}}},
		{"Flag", "one --three=THREE", hookContext{true, []string{"before:", "after:THREE"}}},
		{"ArgAndFlag", "one two --three=THREE", hookContext{true, []string{"before:", "before:", "after:two", "after:THREE"}}},
	}

	var cli struct {
		One hookCmd `cmd:""`
	}

	ctx := &hookContext{}
	p := mustNew(t, &cli, kong.Bind(ctx))

	for _, test := range tests {
		*ctx = hookContext{}
		cli.One = hookCmd{}
		t.Run(test.name, func(t *testing.T) {
			_, err := p.Parse(strings.Split(test.input, " "))
			require.NoError(t, err)
			require.Equal(t, &test.values, ctx)
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

func TestDuplicateSliceAccumulates(t *testing.T) {
	var cli struct {
		Flag []int
	}

	args := []string{"--flag=1,2", "--flag=3,4"}
	_, err := mustNew(t, &cli).Parse(args)
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3, 4}, cli.Flag)
}

func TestMapFlag(t *testing.T) {
	var cli struct {
		Set map[string]int
	}
	_, err := mustNew(t, &cli).Parse([]string{"--set", "a=10", "--set", "b=20"})
	require.NoError(t, err)
	require.Equal(t, map[string]int{"a": 10, "b": 20}, cli.Set)
}

func TestMapFlagWithSliceValue(t *testing.T) {
	var cli struct {
		Set map[string][]int
	}
	_, err := mustNew(t, &cli).Parse([]string{"--set", "a=1,2", "--set", "b=3"})
	require.NoError(t, err)
	require.Equal(t, map[string][]int{"a": {1, 2}, "b": {3}}, cli.Set)
}

type embeddedFlags struct {
	Embedded string
}

func TestEmbeddedStruct(t *testing.T) {
	var cli struct {
		embeddedFlags
		NotEmbedded string
	}

	_, err := mustNew(t, &cli).Parse([]string{"--embedded=moo", "--not-embedded=foo"})
	require.NoError(t, err)
	require.Equal(t, "moo", cli.Embedded)
	require.Equal(t, "foo", cli.NotEmbedded)
}

func TestSliceWithDisabledSeparator(t *testing.T) {
	var cli struct {
		Flag []string `sep:"none"`
	}
	_, err := mustNew(t, &cli).Parse([]string{"--flag=a,b", "--flag=b,c"})
	require.NoError(t, err)
	require.Equal(t, []string{"a,b", "b,c"}, cli.Flag)
}

func TestMultilineMessage(t *testing.T) {
	w := &bytes.Buffer{}
	var cli struct{}
	p := mustNew(t, &cli, kong.Writers(w, w))
	p.Printf("hello\nworld")
	require.Equal(t, "test: hello\n      world\n", w.String())
}

type cmdWithRun struct {
	Arg string `arg:""`
}

func (c *cmdWithRun) Run(key string) error {
	c.Arg += key
	if key == "ERROR" {
		return fmt.Errorf("ERROR")
	}
	return nil
}

type parentCmdWithRun struct {
	Flag       string
	SubCommand struct {
		Arg string `arg:""`
	} `cmd:""`
}

func (p *parentCmdWithRun) Run(key string) error {
	p.SubCommand.Arg += key
	return nil
}

type grammarWithRun struct {
	One   cmdWithRun       `cmd:""`
	Two   cmdWithRun       `cmd:""`
	Three parentCmdWithRun `cmd:""`
}

func TestRun(t *testing.T) {
	cli := &grammarWithRun{}
	p := mustNew(t, cli)

	ctx, err := p.Parse([]string{"one", "two"})
	require.NoError(t, err)
	err = ctx.Run("hello")
	require.NoError(t, err)
	require.Equal(t, "twohello", cli.One.Arg)

	ctx, err = p.Parse([]string{"two", "three"})
	require.NoError(t, err)
	err = ctx.Run("ERROR")
	require.Error(t, err)

	ctx, err = p.Parse([]string{"three", "sub-command", "arg"})
	require.NoError(t, err)
	err = ctx.Run("ping")
	require.NoError(t, err)
	require.Equal(t, "argping", cli.Three.SubCommand.Arg)
}

func TestInterpolationIntoModel(t *testing.T) {
	var cli struct {
		Flag    string `default:"${default}" help:"Help, I need ${somebody}" enum:"${enum}"`
		EnumRef string `enum:"a,b" required:"" help:"One of ${enum}"`
	}
	_, err := kong.New(&cli)
	require.Error(t, err)
	p, err := kong.New(&cli, kong.Vars{
		"default":  "Some default value.",
		"somebody": "chickens!",
		"enum":     "a,b,c,d",
	})
	require.NoError(t, err)
	flag := p.Model.Flags[1]
	flag2 := p.Model.Flags[2]
	require.Equal(t, "Some default value.", flag.Default)
	require.Equal(t, "Help, I need chickens!", flag.Help)
	require.Equal(t, map[string]bool{"a": true, "b": true, "c": true, "d": true}, flag.EnumMap())
	require.Equal(t, "One of a,b", flag2.Help)
}

func TestErrorMissingArgs(t *testing.T) {
	var cli struct {
		One string `arg:""`
		Two string `arg:""`
	}

	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	require.Error(t, err)
	require.Equal(t, "expected \"<one> <two>\"", err.Error())
}

func TestBoolOverride(t *testing.T) {
	var cli struct {
		Flag bool `default:"true"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--flag=false"})
	require.NoError(t, err)
	_, err = p.Parse([]string{"--flag", "false"})
	require.Error(t, err)
}

func TestAnonymousPrefix(t *testing.T) {
	type Anonymous struct {
		Flag string
	}
	var cli struct {
		Anonymous `prefix:"anon-"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--anon-flag=moo"})
	require.NoError(t, err)
	require.Equal(t, "moo", cli.Flag)
}

type TestInterface interface {
	SomeMethod()
}

type TestImpl struct {
	Flag string
}

func (t *TestImpl) SomeMethod() {}

func TestEmbedInterface(t *testing.T) {
	type CLI struct {
		SomeFlag string
		TestInterface
	}
	cli := &CLI{TestInterface: &TestImpl{}}
	p := mustNew(t, cli)
	_, err := p.Parse([]string{"--some-flag=foo", "--flag=yes"})
	require.NoError(t, err)
	require.Equal(t, "foo", cli.SomeFlag)
	require.Equal(t, "yes", cli.TestInterface.(*TestImpl).Flag)
}

func TestExcludedField(t *testing.T) {
	var cli struct {
		Flag     string
		Excluded string `kong:"-"`
	}

	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--flag=foo"})
	require.NoError(t, err)
	_, err = p.Parse([]string{"--excluded=foo"})
	require.Error(t, err)
}

func TestUnnamedFieldEmbeds(t *testing.T) {
	type Embed struct {
		Flag string
	}
	var cli struct {
		One Embed `prefix:"one-" embed:""`
		Two Embed `prefix:"two-" embed:""`
	}
	buf := &strings.Builder{}
	p := mustNew(t, &cli, kong.Writers(buf, buf), kong.Exit(func(int) {}))
	_, err := p.Parse([]string{"--help"})
	require.NoError(t, err)
	require.Contains(t, buf.String(), `--one-flag=STRING`)
	require.Contains(t, buf.String(), `--two-flag=STRING`)
}

func TestHooksCalledForDefault(t *testing.T) {
	var cli struct {
		Flag hookValue `default:"default"`
	}

	ctx := &hookContext{}
	_, err := mustNew(t, &cli, kong.Bind(ctx)).Parse(nil)
	require.NoError(t, err)
	require.Equal(t, "default", string(cli.Flag))
	require.Equal(t, []string{"before:default", "after:default"}, ctx.values)
}

func TestEnum(t *testing.T) {
	var cli struct {
		Flag string `enum:"a,b,c" required:""`
	}
	_, err := mustNew(t, &cli).Parse([]string{"--flag", "d"})
	require.EqualError(t, err, "--flag must be one of \"a\",\"b\",\"c\" but got \"d\"")
}

type commandWithHook struct {
	value string
}

func (c *commandWithHook) AfterApply(cli *cliWithHook) error {
	c.value = cli.Flag
	return nil
}

type cliWithHook struct {
	Flag    string
	Command commandWithHook `cmd:""`
}

func (c *cliWithHook) AfterApply(ctx *kong.Context) error {
	ctx.Bind(c)
	return nil
}

func TestParentBindings(t *testing.T) {
	cli := &cliWithHook{}
	_, err := mustNew(t, cli).Parse([]string{"command", "--flag=foo"})
	require.NoError(t, err)
	require.Equal(t, "foo", cli.Command.value)
}

func TestNumericParamErrors(t *testing.T) {
	var cli struct {
		Name string
	}
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{"--name", "-10"})
	require.EqualError(t, err, `--name: expected string value but got "-10" (short flag); perhaps try --name="-10"?`)
}

func TestDefaultValueIsHyphen(t *testing.T) {
	var cli struct {
		Flag string `default:"-"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, "-", cli.Flag)
}

func TestDefaultEnumValidated(t *testing.T) {
	var cli struct {
		Flag string `default:"invalid" enum:"valid"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	require.EqualError(t, err, "--flag must be one of \"valid\" but got \"invalid\"")
}

func TestEnvarEnumValidated(t *testing.T) {
	restore := tempEnv(map[string]string{
		"FLAG": "invalid",
	})
	defer restore()
	var cli struct {
		Flag string `env:"FLAG" required:"" enum:"valid"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	require.EqualError(t, err, "--flag must be one of \"valid\" but got \"invalid\"")
}

func TestXor(t *testing.T) {
	var cli struct {
		Hello bool   `xor:"another"`
		One   bool   `xor:"group"`
		Two   string `xor:"group"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--hello", "--one", "--two=hi"})
	require.EqualError(t, err, "--one and --two can't be used together")

	p = mustNew(t, &cli)
	_, err = p.Parse([]string{"--one", "--hello"})
	require.NoError(t, err)
}

func TestXorChild(t *testing.T) {
	var cli struct {
		One bool `xor:"group"`
		Cmd struct {
			Two   string `xor:"group"`
			Three string `xor:"group"`
		} `cmd`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--one", "cmd", "--two=hi"})
	require.NoError(t, err)

	p = mustNew(t, &cli)
	_, err = p.Parse([]string{"--two=hi", "cmd", "--three"})
	require.Error(t, err, "--two and --three can't be used together")
}

func TestMultiXor(t *testing.T) {
	var cli struct {
		Hello bool   `xor:"one,two"`
		One   bool   `xor:"one"`
		Two   string `xor:"two"`
	}

	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--hello", "--one"})
	require.EqualError(t, err, "--hello and --one can't be used together")

	p = mustNew(t, &cli)
	_, err = p.Parse([]string{"--hello", "--two=foo"})
	require.EqualError(t, err, "--hello and --two can't be used together")
}

func TestXorRequired(t *testing.T) {
	var cli struct {
		One   bool `xor:"one,two" required:""`
		Two   bool `xor:"one" required:""`
		Three bool `xor:"two" required:""`
		Four  bool `required:""`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--one"})
	require.EqualError(t, err, "missing flags: --four")

	p = mustNew(t, &cli)
	_, err = p.Parse([]string{"--two"})
	require.EqualError(t, err, "missing flags: --four, --one or --three")

	p = mustNew(t, &cli)
	_, err = p.Parse([]string{})
	require.EqualError(t, err, "missing flags: --four, --one or --three, --one or --two")
}

func TestXorRequiredMany(t *testing.T) {
	var cli struct {
		One   bool `xor:"one" required:""`
		Two   bool `xor:"one" required:""`
		Three bool `xor:"one" required:""`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--one"})
	require.NoError(t, err)

	p = mustNew(t, &cli)
	_, err = p.Parse([]string{})
	require.EqualError(t, err, "missing flags: --one or --two or --three")
}

func TestEnumSequence(t *testing.T) {
	var cli struct {
		State []string `enum:"a,b,c" default:"a"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, []string{"a"}, cli.State)
}

func TestIssue40EnumAcrossCommands(t *testing.T) {
	var cli struct {
		One struct {
			OneArg string `arg:"" required:""`
		} `cmd:""`
		Two struct {
			TwoArg string `arg:"" enum:"a,b,c" required:"" env:"FOO"`
		} `cmd:""`
		Three struct {
			ThreeArg string `arg:"" optional:"" default:"a" enum:"a,b,c"`
		} `cmd:""`
	}

	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"one", "two"})
	require.NoError(t, err)
	_, err = p.Parse([]string{"two", "d"})
	require.Error(t, err)
	_, err = p.Parse([]string{"three", "d"})
	require.Error(t, err)
	_, err = p.Parse([]string{"three", "c"})
	require.NoError(t, err)
}

func TestIssue179(t *testing.T) {
	type A struct {
		Enum string `required:"" enum:"1,2"`
	}

	type B struct{}

	var root struct {
		A A `cmd`
		B B `cmd`
	}

	p := mustNew(t, &root)
	_, err := p.Parse([]string{"b"})
	require.NoError(t, err)
}

func TestIssue153(t *testing.T) {
	type LsCmd struct {
		Paths []string `arg required name:"path" help:"Paths to list." env:"CMD_PATHS"`
	}

	var cli struct {
		Debug bool `help:"Enable debug mode."`

		Ls LsCmd `cmd help:"List paths."`
	}

	p, revert := newEnvParser(t, &cli, envMap{
		"CMD_PATHS": "hello",
	})
	defer revert()
	_, err := p.Parse([]string{"ls"})
	require.NoError(t, err)
	require.Equal(t, []string{"hello"}, cli.Ls.Paths)
}

func TestEnumArg(t *testing.T) {
	var cli struct {
		Nested struct {
			One string `arg:"" enum:"a,b,c" required:""`
			Two string `arg:""`
		} `cmd:""`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"nested", "a", "b"})
	require.NoError(t, err)
	require.Equal(t, "a", cli.Nested.One)
	require.Equal(t, "b", cli.Nested.Two)
}

func TestDefaultCommand(t *testing.T) {
	var cli struct {
		One struct{} `cmd:"" default:"1"`
		Two struct{} `cmd:""`
	}
	p := mustNew(t, &cli)
	ctx, err := p.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "one", ctx.Command())
}

func TestMultipleDefaultCommands(t *testing.T) {
	var cli struct {
		One struct{} `cmd:"" default:"1"`
		Two struct{} `cmd:"" default:"1"`
	}
	_, err := kong.New(&cli)
	require.EqualError(t, err, "<anonymous struct>.Two: can't have more than one default command under  <command>")
}

func TestDefaultCommandWithSubCommand(t *testing.T) {
	var cli struct {
		One struct {
			Two struct{} `cmd:""`
		} `cmd:"" default:"1"`
	}
	_, err := kong.New(&cli)
	require.EqualError(t, err, "<anonymous struct>.One: default command one <command> must not have subcommands or arguments")
}

func TestDefaultCommandWithAllowedSubCommand(t *testing.T) {
	var cli struct {
		One struct {
			Two struct{} `cmd:""`
		} `cmd:"" default:"withargs"`
	}
	p := mustNew(t, &cli)
	ctx, err := p.Parse([]string{"two"})
	require.NoError(t, err)
	require.Equal(t, "one two", ctx.Command())
}

func TestDefaultCommandWithArgument(t *testing.T) {
	var cli struct {
		One struct {
			Arg string `arg:""`
		} `cmd:"" default:"1"`
	}
	_, err := kong.New(&cli)
	require.EqualError(t, err, "<anonymous struct>.One: default command one <arg> must not have subcommands or arguments")
}

func TestDefaultCommandWithAllowedArgument(t *testing.T) {
	var cli struct {
		One struct {
			Arg  string `arg:""`
			Flag string
		} `cmd:"" default:"withargs"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"arg", "--flag=value"})
	require.NoError(t, err)
	require.Equal(t, "arg", cli.One.Arg)
	require.Equal(t, "value", cli.One.Flag)
}

func TestDefaultCommandWithBranchingArgument(t *testing.T) {
	var cli struct {
		One struct {
			Two struct {
				Two string `arg:""`
			} `arg:""`
		} `cmd:"" default:"1"`
	}
	_, err := kong.New(&cli)
	require.EqualError(t, err, "<anonymous struct>.One: default command one <command> must not have subcommands or arguments")
}

func TestDefaultCommandWithAllowedBranchingArgument(t *testing.T) {
	var cli struct {
		One struct {
			Two struct {
				Two  string `arg:""`
				Flag string
			} `arg:""`
		} `cmd:"" default:"withargs"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"arg", "--flag=value"})
	require.NoError(t, err)
	require.Equal(t, "arg", cli.One.Two.Two)
	require.Equal(t, "value", cli.One.Two.Flag)
}

func TestDefaultCommandPrecedence(t *testing.T) {
	var cli struct {
		Two struct {
			Arg  string `arg:""`
			Flag bool
		} `cmd:"" default:"withargs"`
		One struct{} `cmd:""`
	}
	p := mustNew(t, &cli)

	// A named command should take precedence over a default command with arg
	ctx, err := p.Parse([]string{"one"})
	require.NoError(t, err)
	require.Equal(t, "one", ctx.Command())

	// An explicitly named command with arg should parse, even if labeled default:"witharg"
	ctx, err = p.Parse([]string{"two", "arg"})
	require.NoError(t, err)
	require.Equal(t, "two <arg>", ctx.Command())

	// An arg to a default command that does not match another command should select the default
	ctx, err = p.Parse([]string{"arg"})
	require.NoError(t, err)
	require.Equal(t, "two <arg>", ctx.Command())

	// A flag on a default command should not be valid on a sibling command
	_, err = p.Parse([]string{"one", "--flag"})
	require.EqualError(t, err, "unknown flag --flag")
}

func TestLoneHpyhen(t *testing.T) {
	var cli struct {
		Flag string
		Arg  string `arg:"" optional:""`
	}
	p := mustNew(t, &cli)

	_, err := p.Parse([]string{"-"})
	require.NoError(t, err)
	require.Equal(t, "-", cli.Arg)

	_, err = p.Parse([]string{"--flag", "-"})
	require.NoError(t, err)
	require.Equal(t, "-", cli.Flag)
}

func TestPlugins(t *testing.T) {
	var pluginOne struct {
		One string
	}
	var pluginTwo struct {
		Two string
	}
	var cli struct {
		Base string
		kong.Plugins
	}
	cli.Plugins = kong.Plugins{&pluginOne, &pluginTwo}

	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--base=base", "--one=one", "--two=two"})
	require.NoError(t, err)
	require.Equal(t, "base", cli.Base)
	require.Equal(t, "one", pluginOne.One)
	require.Equal(t, "two", pluginTwo.Two)
}

type validateCmd struct{}

func (v *validateCmd) Validate() error { return errors.New("cmd error") }

type validateCli struct {
	Cmd validateCmd `cmd:""`
}

func (v *validateCli) Validate() error { return errors.New("app error") }

type validateFlag string

func (v *validateFlag) Validate() error { return errors.New("flag error") }

func TestValidateApp(t *testing.T) {
	cli := validateCli{}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{})
	require.EqualError(t, err, "test: app error")
}

func TestValidateCmd(t *testing.T) {
	cli := struct {
		Cmd validateCmd `cmd:""`
	}{}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"cmd"})
	require.EqualError(t, err, "cmd: cmd error")
}

func TestValidateFlag(t *testing.T) {
	cli := struct {
		Flag validateFlag
	}{}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--flag=one"})
	require.EqualError(t, err, "--flag: flag error")
}

func TestValidateArg(t *testing.T) {
	cli := struct {
		Arg validateFlag `arg:""`
	}{}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"one"})
	require.EqualError(t, err, "<arg>: flag error")
}

func TestPointers(t *testing.T) {
	cli := struct {
		Mapped *mappedValue
		JSON   *jsonUnmarshalerValue
	}{}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--mapped=mapped", "--json=\"foo\""})
	require.NoError(t, err)
	require.NotNil(t, cli.Mapped)
	require.Equal(t, "mapped", cli.Mapped.decoded)
	require.NotNil(t, cli.JSON)
	require.Equal(t, "FOO", string(*cli.JSON))
}

type dynamicCommand struct {
	Flag string

	ran bool
}

func (d *dynamicCommand) Run() error {
	d.ran = true
	return nil
}

func TestDynamicCommands(t *testing.T) {
	cli := struct {
		One struct{} `cmd:"one"`
	}{}
	help := &strings.Builder{}
	two := &dynamicCommand{}
	three := &dynamicCommand{}
	p := mustNew(t, &cli,
		kong.DynamicCommand("two", "", "", &two),
		kong.DynamicCommand("three", "", "", three, "hidden"),
		kong.Writers(help, help),
		kong.Exit(func(int) {}))
	kctx, err := p.Parse([]string{"two", "--flag=flag"})
	require.NoError(t, err)
	require.Equal(t, "flag", two.Flag)
	require.False(t, two.ran)
	err = kctx.Run()
	require.NoError(t, err)
	require.True(t, two.ran)

	_, err = p.Parse([]string{"--help"})
	require.EqualError(t, err, `expected one of "one",  "two"`)
	require.NotContains(t, help.String(), "three", help.String())
}

func TestDuplicateShortflags(t *testing.T) {
	cli := struct {
		Flag1 bool `short:"t"`
		Flag2 bool `short:"t"`
	}{}
	_, err := kong.New(&cli)
	require.EqualError(t, err, "<anonymous struct>.Flag2: duplicate short flag -t")
}

func TestDuplicateNestedShortFlags(t *testing.T) {
	cli := struct {
		Flag1 bool `short:"t"`
		Cmd   struct {
			Flag2 bool `short:"t"`
		} `cmd:""`
	}{}
	_, err := kong.New(&cli)
	require.EqualError(t, err, "<anonymous struct>.Flag2: duplicate short flag -t")
}

func TestHydratePointerCommands(t *testing.T) {
	type cmd struct {
		Flag bool
	}

	var cli struct {
		Cmd *cmd `cmd:""`
	}

	k := mustNew(t, &cli)
	_, err := k.Parse([]string{"cmd", "--flag"})
	require.NoError(t, err)
	require.Equal(t, &cmd{Flag: true}, cli.Cmd)
}

type testIgnoreFields struct {
	Foo struct {
		Bar bool
		Sub struct {
			SubFlag1     bool `kong:"name=subflag1"`
			XXX_SubFlag2 bool `kong:"name=subflag2"`
		} `kong:"cmd"`
	} `kong:"cmd"`
	XXX_Baz struct {
		Boo bool
	} `kong:"cmd,name=baz"`
}

func TestIgnoreRegex(t *testing.T) {
	r, err := regexp.Compile("^XXX_")
	require.NoError(t, err)

	cli := testIgnoreFields{}

	k, err := kong.New(&cli, kong.IgnoreFieldsRegex(r))
	require.NoError(t, err)

	_, err = k.Parse([]string{"foo", "sub"})
	require.NoError(t, err)

	_, err = k.Parse([]string{"foo", "sub", "--subflag1"})
	require.NoError(t, err)

	_, err = k.Parse([]string{"foo", "sub", "--subflag2"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown flag --subflag2")

	_, err = k.Parse([]string{"baz"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected argument baz")
}

// Verify that passing a nil regex will work
func TestIgnoreRegexNil(t *testing.T) {
	cli := testIgnoreFields{}

	k, err := kong.New(&cli, kong.IgnoreFieldsRegex(nil))
	require.NoError(t, err)

	_, err = k.Parse([]string{"foo", "sub", "--subflag1", "--subflag2"})
	require.NoError(t, err)

	_, err = k.Parse([]string{"baz"})
	require.NoError(t, err)
}

type optionWithErr struct{}

func (o *optionWithErr) Apply(k *kong.Kong) error {
	return errors.New("option returned err")
}

func TestOptionReturnsErr(t *testing.T) {
	cli := struct {
		Test bool `flag"`
	}{}

	optWithError := &optionWithErr{}

	_, err := kong.New(cli, optWithError)
	require.Error(t, err)
	require.Equal(t, "option returned err", err.Error())
}
