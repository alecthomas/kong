package kong_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

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
	// nolint: govet
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
	// nolint: govet
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
		// nolint: scopelint
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
		EnumRef string `enum:"a,b" help:"One of ${enum}"`
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
		Flag string `enum:"a,b,c"`
	}
	_, err := mustNew(t, &cli).Parse([]string{"--flag", "d"})
	require.EqualError(t, err, "--flag=STRING must be one of a,b,c but got \"\"")
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
	require.EqualError(t, err, `expected string value but got "-10" (short flag)`)
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
	require.Error(t, err)
}
