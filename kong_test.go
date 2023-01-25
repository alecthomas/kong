package kong_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/repr"

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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
	assert.Equal(t, "user create <id> <first> <last>", ctx.Command())
	t.Run("Missing", func(t *testing.T) {
		_, err := p.Parse([]string{"user", "create", "10"})
		assert.Error(t, err)
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
	assert.NoError(t, err)
	assert.Equal(t, 10, cli.User.ID.ID)
	assert.Equal(t, "user <id> delete", ctx.Command())
	t.Run("Missing", func(t *testing.T) {
		_, err = p.Parse([]string{"user"})
		assert.Error(t, err)
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
	assert.NoError(t, err)
	assert.Equal(t, "", cli.Flag)
	assert.Equal(t, "default", cli.FlagWithDefault)
}

func TestFlagSlice(t *testing.T) {
	var cli struct {
		Slice []int
	}
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{"--slice=1,2,3"})
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, cli.Slice)
}

func TestFlagSliceWithSeparator(t *testing.T) {
	var cli struct {
		Slice []string
	}
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{`--slice=a\,b,c`})
	assert.NoError(t, err)
	assert.Equal(t, []string{"a,b", "c"}, cli.Slice)
}

func TestArgSlice(t *testing.T) {
	var cli struct {
		Slice []int `arg`
		Flag  bool
	}
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{"1", "2", "3", "--flag"})
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, cli.Slice)
	assert.Equal(t, true, cli.Flag)
}

func TestArgSliceWithSeparator(t *testing.T) {
	var cli struct {
		Slice []string `arg`
		Flag  bool
	}
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{"a,b", "c", "--flag"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"a,b", "c"}, cli.Slice)
	assert.Equal(t, true, cli.Flag)
}

func TestUnsupportedFieldErrors(t *testing.T) {
	var cli struct {
		Keys struct{}
	}
	_, err := kong.New(&cli)
	assert.Error(t, err)
}

func TestMatchingArgField(t *testing.T) {
	var cli struct {
		ID struct {
			NotID int `kong:"arg"`
		} `kong:"arg"`
	}

	_, err := kong.New(&cli)
	assert.Error(t, err)
}

func TestCantMixPositionalAndBranches(t *testing.T) {
	var cli struct {
		Arg     string `kong:"arg"`
		Command struct {
		} `kong:"cmd"`
	}
	_, err := kong.New(&cli)
	assert.Error(t, err)
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
	assert.NoError(t, err)
	assert.Equal(t, "moo", cli.Flag1)
	assert.Equal(t, true, cli.Command1.Flag2)
}

func TestRequiredFlag(t *testing.T) {
	var cli struct {
		Flag string `kong:"required"`
	}

	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{})
	assert.Error(t, err)
}

func TestOptionalArg(t *testing.T) {
	var cli struct {
		Arg string `kong:"arg,optional"`
	}

	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{})
	assert.NoError(t, err)
}

func TestOptionalArgWithDefault(t *testing.T) {
	var cli struct {
		Arg string `kong:"arg,optional,default='moo'"`
	}

	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "moo", cli.Arg)
}

func TestArgWithDefaultIsOptional(t *testing.T) {
	var cli struct {
		Arg string `kong:"arg,default='moo'"`
	}

	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "moo", cli.Arg)
}

func TestRequiredArg(t *testing.T) {
	var cli struct {
		Arg string `kong:"arg"`
	}

	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{})
	assert.Error(t, err)
}

func TestInvalidRequiredAfterOptional(t *testing.T) {
	var cli struct {
		ID   int    `kong:"arg,optional"`
		Name string `kong:"arg"`
	}

	_, err := kong.New(&cli)
	assert.Error(t, err)
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
		assert.NoError(t, err)
		assert.Equal(t, "gak", cli.Name.Name)
		assert.Equal(t, true, cli.Name.Enabled)
	})

	t.Run("WithoutFlag", func(t *testing.T) {
		_, err := parser.Parse([]string{"gak"})
		assert.NoError(t, err)
		assert.Equal(t, "gak", cli.Name.Name)
	})

	t.Run("WithNothing", func(t *testing.T) {
		_, err := parser.Parse([]string{})
		assert.NoError(t, err)
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
		assert.NoError(t, err)
		assert.Equal(t, "gak", cli.Name)
		assert.Equal(t, 5, cli.ID)
	})

	t.Run("ExtraOptional", func(t *testing.T) {
		_, err := parser.Parse([]string{"gak"})
		assert.NoError(t, err)
		assert.Equal(t, "gak", cli.Name)
	})
}

func TestInvalidDefaultErrors(t *testing.T) {
	var cli struct {
		Flag int `kong:"default='foo'"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	assert.Error(t, err)
}

func TestCommandMissingTagIsInvalid(t *testing.T) {
	var cli struct {
		One struct{}
	}
	_, err := kong.New(&cli)
	assert.Error(t, err)
}

func TestDuplicateFlag(t *testing.T) {
	var cli struct {
		Flag bool
		Cmd  struct {
			Flag bool
		} `kong:"cmd"`
	}
	_, err := kong.New(&cli)
	assert.Error(t, err)
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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
	assert.Error(t, ctx.Error)
	assert.Equal(t, "one", ctx.Command())
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
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, cli.Cmd.Flag)

			err = kctx.Run()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, cli.Cmd.Flag)
			assert.True(t, cli.Cmd.ran)
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
	assert.NoError(t, err)
	assert.Equal(t, true, cli.Cmd.Flag)
	assert.Equal(t, "none", cli.Cmd.NoFlag)
}

func TestInvalidNegatedNonBool(t *testing.T) {
	var cli struct {
		Cmd struct {
			Flag string `kong:"negatable"`
		} `kong:"cmd"`
	}

	_, err := kong.New(&cli)
	assert.Error(t, err)
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
			assert.NoError(t, err)
			assert.Equal(t, &test.values, ctx)
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
	assert.NoError(t, err)
	assert.True(t, cli.Bool)
	assert.Equal(t, "hello", cli.String)
}

func TestDuplicateFlagChoosesLast(t *testing.T) {
	var cli struct {
		Flag int
	}

	_, err := mustNew(t, &cli).Parse([]string{"--flag=1", "--flag=2"})
	assert.NoError(t, err)
	assert.Equal(t, 2, cli.Flag)
}

func TestDuplicateSliceAccumulates(t *testing.T) {
	var cli struct {
		Flag []int
	}

	args := []string{"--flag=1,2", "--flag=3,4"}
	_, err := mustNew(t, &cli).Parse(args)
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4}, cli.Flag)
}

func TestMapFlag(t *testing.T) {
	var cli struct {
		Set map[string]int
	}
	_, err := mustNew(t, &cli).Parse([]string{"--set", "a=10", "--set", "b=20"})
	assert.NoError(t, err)
	assert.Equal(t, map[string]int{"a": 10, "b": 20}, cli.Set)
}

func TestMapFlagWithSliceValue(t *testing.T) {
	var cli struct {
		Set map[string][]int
	}
	_, err := mustNew(t, &cli).Parse([]string{"--set", "a=1,2", "--set", "b=3"})
	assert.NoError(t, err)
	assert.Equal(t, map[string][]int{"a": {1, 2}, "b": {3}}, cli.Set)
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
	assert.NoError(t, err)
	assert.Equal(t, "moo", cli.Embedded)
	assert.Equal(t, "foo", cli.NotEmbedded)
}

func TestSliceWithDisabledSeparator(t *testing.T) {
	var cli struct {
		Flag []string `sep:"none"`
	}
	_, err := mustNew(t, &cli).Parse([]string{"--flag=a,b", "--flag=b,c"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"a,b", "b,c"}, cli.Flag)
}

func TestMultilineMessage(t *testing.T) {
	w := &bytes.Buffer{}
	var cli struct{}
	p := mustNew(t, &cli, kong.Writers(w, w))
	p.Printf("hello\nworld")
	assert.Equal(t, "test: hello\n      world\n", w.String())
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
	assert.NoError(t, err)
	err = ctx.Run("hello")
	assert.NoError(t, err)
	assert.Equal(t, "twohello", cli.One.Arg)

	ctx, err = p.Parse([]string{"two", "three"})
	assert.NoError(t, err)
	err = ctx.Run("ERROR")
	assert.Error(t, err)

	ctx, err = p.Parse([]string{"three", "sub-command", "arg"})
	assert.NoError(t, err)
	err = ctx.Run("ping")
	assert.NoError(t, err)
	assert.Equal(t, "argping", cli.Three.SubCommand.Arg)
}

func TestInterpolationIntoModel(t *testing.T) {
	var cli struct {
		Flag    string `default:"${default_value}" help:"Help, I need ${somebody}" enum:"${enum}"`
		EnumRef string `enum:"a,b" required:"" help:"One of ${enum}"`
		EnvRef  string `env:"${env}" help:"God ${env}"`
	}
	_, err := kong.New(&cli)
	assert.Error(t, err)
	p, err := kong.New(&cli, kong.Vars{
		"default_value": "Some default value.",
		"somebody":      "chickens!",
		"enum":          "a,b,c,d",
		"env":           "SAVE_THE_QUEEN",
	})
	assert.NoError(t, err)
	assert.Equal(t, 4, len(p.Model.Flags))
	flag := p.Model.Flags[1]
	flag2 := p.Model.Flags[2]
	flag3 := p.Model.Flags[3]
	assert.Equal(t, "Some default value.", flag.Default)
	assert.Equal(t, "Help, I need chickens!", flag.Help)
	assert.Equal(t, map[string]bool{"a": true, "b": true, "c": true, "d": true}, flag.EnumMap())
	assert.Equal(t, []string{"a", "b", "c", "d"}, flag.EnumSlice())
	assert.Equal(t, "One of a,b", flag2.Help)
	assert.Equal(t, "SAVE_THE_QUEEN", flag3.Env)
	assert.Equal(t, "God SAVE_THE_QUEEN", flag3.Help)
}

func TestIssue244(t *testing.T) {
	type Config struct {
		Project string `short:"p" env:"CI_PROJECT_ID" help:"Environment variable: ${env}"`
	}
	w := &strings.Builder{}
	k := mustNew(t, &Config{}, kong.Exit(func(int) {}), kong.Writers(w, w))
	_, err := k.Parse([]string{"--help"})
	assert.NoError(t, err)
	assert.Contains(t, w.String(), `Environment variable: CI_PROJECT_ID`)
}

func TestErrorMissingArgs(t *testing.T) {
	var cli struct {
		One string `arg:""`
		Two string `arg:""`
	}

	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	assert.Error(t, err)
	assert.Equal(t, "expected \"<one> <two>\"", err.Error())
}

func TestBoolOverride(t *testing.T) {
	var cli struct {
		Flag bool `default:"true"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--flag=false"})
	assert.NoError(t, err)
	_, err = p.Parse([]string{"--flag", "false"})
	assert.Error(t, err)
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
	assert.NoError(t, err)
	assert.Equal(t, "moo", cli.Flag)
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
	assert.NoError(t, err)
	assert.Equal(t, "foo", cli.SomeFlag)
	assert.Equal(t, "yes", cli.TestInterface.(*TestImpl).Flag) // nolint
}

func TestExcludedField(t *testing.T) {
	var cli struct {
		Flag     string
		Excluded string `kong:"-"`
	}

	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--flag=foo"})
	assert.NoError(t, err)
	_, err = p.Parse([]string{"--excluded=foo"})
	assert.Error(t, err)
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
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), `--one-flag=STRING`)
	assert.Contains(t, buf.String(), `--two-flag=STRING`)
}

func TestHooksCalledForDefault(t *testing.T) {
	var cli struct {
		Flag hookValue `default:"default"`
	}

	ctx := &hookContext{}
	_, err := mustNew(t, &cli, kong.Bind(ctx)).Parse(nil)
	assert.NoError(t, err)
	assert.Equal(t, "default", string(cli.Flag))
	assert.Equal(t, []string{"before:default", "after:default"}, ctx.values)
}

func TestEnum(t *testing.T) {
	var cli struct {
		Flag string `enum:"a,b,c" required:""`
	}
	_, err := mustNew(t, &cli).Parse([]string{"--flag", "d"})
	assert.EqualError(t, err, "--flag must be one of \"a\",\"b\",\"c\" but got \"d\"")
}

func TestEnumMeaningfulOrder(t *testing.T) {
	var cli struct {
		Flag string `enum:"first,second,third,fourth,fifth" required:""`
	}
	_, err := mustNew(t, &cli).Parse([]string{"--flag", "sixth"})
	assert.EqualError(t, err, "--flag must be one of \"first\",\"second\",\"third\",\"fourth\",\"fifth\" but got \"sixth\"")
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
	assert.NoError(t, err)
	assert.Equal(t, "foo", cli.Command.value)
}

func TestNumericParamErrors(t *testing.T) {
	var cli struct {
		Name string
	}
	parser := mustNew(t, &cli)
	_, err := parser.Parse([]string{"--name", "-10"})
	assert.EqualError(t, err, `--name: expected string value but got "-10" (short flag); perhaps try --name="-10"?`)
}

func TestDefaultValueIsHyphen(t *testing.T) {
	var cli struct {
		Flag string `default:"-"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	assert.NoError(t, err)
	assert.Equal(t, "-", cli.Flag)
}

func TestDefaultEnumValidated(t *testing.T) {
	var cli struct {
		Flag string `default:"invalid" enum:"valid"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	assert.EqualError(t, err, "--flag must be one of \"valid\" but got \"invalid\"")
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
	assert.EqualError(t, err, "--flag must be one of \"valid\" but got \"invalid\"")
}

func TestXor(t *testing.T) {
	var cli struct {
		Hello bool   `xor:"another"`
		One   bool   `xor:"group"`
		Two   string `xor:"group"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--hello", "--one", "--two=hi"})
	assert.EqualError(t, err, "--one and --two can't be used together")

	p = mustNew(t, &cli)
	_, err = p.Parse([]string{"--one", "--hello"})
	assert.NoError(t, err)
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
	assert.NoError(t, err)

	p = mustNew(t, &cli)
	_, err = p.Parse([]string{"--two=hi", "cmd", "--three"})
	assert.Error(t, err, "--two and --three can't be used together")
}

func TestMultiXor(t *testing.T) {
	var cli struct {
		Hello bool   `xor:"one,two"`
		One   bool   `xor:"one"`
		Two   string `xor:"two"`
	}

	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--hello", "--one"})
	assert.EqualError(t, err, "--hello and --one can't be used together")

	p = mustNew(t, &cli)
	_, err = p.Parse([]string{"--hello", "--two=foo"})
	assert.EqualError(t, err, "--hello and --two can't be used together")
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
	assert.EqualError(t, err, "missing flags: --four")

	p = mustNew(t, &cli)
	_, err = p.Parse([]string{"--two"})
	assert.EqualError(t, err, "missing flags: --four, --one or --three")

	p = mustNew(t, &cli)
	_, err = p.Parse([]string{})
	assert.EqualError(t, err, "missing flags: --four, --one or --three, --one or --two")
}

func TestXorRequiredMany(t *testing.T) {
	var cli struct {
		One   bool `xor:"one" required:""`
		Two   bool `xor:"one" required:""`
		Three bool `xor:"one" required:""`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--one"})
	assert.NoError(t, err)

	p = mustNew(t, &cli)
	_, err = p.Parse([]string{"--three"})
	assert.NoError(t, err)

	p = mustNew(t, &cli)
	_, err = p.Parse([]string{})
	assert.EqualError(t, err, "missing flags: --one or --two or --three")
}

func TestEnumSequence(t *testing.T) {
	var cli struct {
		State []string `enum:"a,b,c" default:"a"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"a"}, cli.State)
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
	assert.NoError(t, err)
	_, err = p.Parse([]string{"two", "d"})
	assert.Error(t, err)
	_, err = p.Parse([]string{"three", "d"})
	assert.Error(t, err)
	_, err = p.Parse([]string{"three", "c"})
	assert.NoError(t, err)
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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
	assert.Equal(t, []string{"hello"}, cli.Ls.Paths)
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
	assert.NoError(t, err)
	assert.Equal(t, "a", cli.Nested.One)
	assert.Equal(t, "b", cli.Nested.Two)
}

func TestDefaultCommand(t *testing.T) {
	var cli struct {
		One struct{} `cmd:"" default:"1"`
		Two struct{} `cmd:""`
	}
	p := mustNew(t, &cli)
	ctx, err := p.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "one", ctx.Command())
}

func TestMultipleDefaultCommands(t *testing.T) {
	var cli struct {
		One struct{} `cmd:"" default:"1"`
		Two struct{} `cmd:"" default:"1"`
	}
	_, err := kong.New(&cli)
	assert.EqualError(t, err, "<anonymous struct>.Two: can't have more than one default command under  <command>")
}

func TestDefaultCommandWithSubCommand(t *testing.T) {
	var cli struct {
		One struct {
			Two struct{} `cmd:""`
		} `cmd:"" default:"1"`
	}
	_, err := kong.New(&cli)
	assert.EqualError(t, err, "<anonymous struct>.One: default command one <command> must not have subcommands or arguments")
}

func TestDefaultCommandWithAllowedSubCommand(t *testing.T) {
	var cli struct {
		One struct {
			Two struct{} `cmd:""`
		} `cmd:"" default:"withargs"`
	}
	p := mustNew(t, &cli)
	ctx, err := p.Parse([]string{"two"})
	assert.NoError(t, err)
	assert.Equal(t, "one two", ctx.Command())
}

func TestDefaultCommandWithArgument(t *testing.T) {
	var cli struct {
		One struct {
			Arg string `arg:""`
		} `cmd:"" default:"1"`
	}
	_, err := kong.New(&cli)
	assert.EqualError(t, err, "<anonymous struct>.One: default command one <arg> must not have subcommands or arguments")
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
	assert.NoError(t, err)
	assert.Equal(t, "arg", cli.One.Arg)
	assert.Equal(t, "value", cli.One.Flag)
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
	assert.EqualError(t, err, "<anonymous struct>.One: default command one <command> must not have subcommands or arguments")
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
	assert.NoError(t, err)
	assert.Equal(t, "arg", cli.One.Two.Two)
	assert.Equal(t, "value", cli.One.Two.Flag)
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
	assert.NoError(t, err)
	assert.Equal(t, "one", ctx.Command())

	// An explicitly named command with arg should parse, even if labeled default:"witharg"
	ctx, err = p.Parse([]string{"two", "arg"})
	assert.NoError(t, err)
	assert.Equal(t, "two <arg>", ctx.Command())

	// An arg to a default command that does not match another command should select the default
	ctx, err = p.Parse([]string{"arg"})
	assert.NoError(t, err)
	assert.Equal(t, "two <arg>", ctx.Command())

	// A flag on a default command should not be valid on a sibling command
	_, err = p.Parse([]string{"one", "--flag"})
	assert.EqualError(t, err, "unknown flag --flag")
}

func TestLoneHpyhen(t *testing.T) {
	var cli struct {
		Flag string
		Arg  string `arg:"" optional:""`
	}
	p := mustNew(t, &cli)

	_, err := p.Parse([]string{"-"})
	assert.NoError(t, err)
	assert.Equal(t, "-", cli.Arg)

	_, err = p.Parse([]string{"--flag", "-"})
	assert.NoError(t, err)
	assert.Equal(t, "-", cli.Flag)
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
	assert.NoError(t, err)
	assert.Equal(t, "base", cli.Base)
	assert.Equal(t, "one", pluginOne.One)
	assert.Equal(t, "two", pluginTwo.Two)
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
	assert.EqualError(t, err, "test: app error")
}

func TestValidateCmd(t *testing.T) {
	cli := struct {
		Cmd validateCmd `cmd:""`
	}{}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"cmd"})
	assert.EqualError(t, err, "cmd: cmd error")
}

func TestValidateFlag(t *testing.T) {
	cli := struct {
		Flag validateFlag
	}{}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--flag=one"})
	assert.EqualError(t, err, "--flag: flag error")
}

func TestValidateArg(t *testing.T) {
	cli := struct {
		Arg validateFlag `arg:""`
	}{}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"one"})
	assert.EqualError(t, err, "<arg>: flag error")
}

func TestPointers(t *testing.T) {
	cli := struct {
		Mapped *mappedValue
		JSON   *jsonUnmarshalerValue
	}{}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--mapped=mapped", "--json=\"foo\""})
	assert.NoError(t, err)
	assert.NotZero(t, cli.Mapped)
	assert.Equal(t, "mapped", cli.Mapped.decoded)
	assert.NotZero(t, cli.JSON)
	assert.Equal(t, "FOO", string(*cli.JSON))
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
	assert.NoError(t, err)
	assert.Equal(t, "flag", two.Flag)
	assert.False(t, two.ran)
	err = kctx.Run()
	assert.NoError(t, err)
	assert.True(t, two.ran)

	_, err = p.Parse([]string{"--help"})
	assert.EqualError(t, err, `expected one of "one",  "two"`)
	assert.NotContains(t, help.String(), "three", help.String())
}

func TestDuplicateShortflags(t *testing.T) {
	cli := struct {
		Flag1 bool `short:"t"`
		Flag2 bool `short:"t"`
	}{}
	_, err := kong.New(&cli)
	assert.EqualError(t, err, "<anonymous struct>.Flag2: duplicate short flag -t")
}

func TestDuplicateNestedShortFlags(t *testing.T) {
	cli := struct {
		Flag1 bool `short:"t"`
		Cmd   struct {
			Flag2 bool `short:"t"`
		} `cmd:""`
	}{}
	_, err := kong.New(&cli)
	assert.EqualError(t, err, "<anonymous struct>.Flag2: duplicate short flag -t")
}

func TestHydratePointerCommandsAndEmbeds(t *testing.T) {
	type cmd struct {
		Flag bool
	}

	type embed struct {
		Embed bool
	}

	var cli struct {
		Cmd   *cmd   `cmd:""`
		Embed *embed `embed:""`
	}

	k := mustNew(t, &cli)
	_, err := k.Parse([]string{"--embed", "cmd", "--flag"})
	assert.NoError(t, err)
	assert.Equal(t, &cmd{Flag: true}, cli.Cmd)
	assert.Equal(t, &embed{Embed: true}, cli.Embed)
}

// nolint
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
	cli := testIgnoreFields{}

	k, err := kong.New(&cli, kong.IgnoreFields(`.*\.XXX_.+`))
	assert.NoError(t, err)

	_, err = k.Parse([]string{"foo", "sub"})
	assert.NoError(t, err)

	_, err = k.Parse([]string{"foo", "sub", "--subflag1"})
	assert.NoError(t, err)

	_, err = k.Parse([]string{"foo", "sub", "--subflag2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown flag --subflag2")

	_, err = k.Parse([]string{"baz"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected argument baz")
}

// Verify that passing a nil regex will work
func TestIgnoreRegexEmpty(t *testing.T) {
	cli := testIgnoreFields{}

	_, err := kong.New(&cli, kong.IgnoreFields(""))
	assert.Error(t, err)
	assert.Contains(t, "regex input cannot be empty", err.Error())
}

type optionWithErr struct{}

func (o *optionWithErr) Apply(k *kong.Kong) error {
	return errors.New("option returned err")
}

func TestOptionReturnsErr(t *testing.T) {
	cli := struct {
		Test bool
	}{}

	optWithError := &optionWithErr{}

	_, err := kong.New(cli, optWithError)
	assert.Error(t, err)
	assert.Equal(t, "option returned err", err.Error())
}

func TestEnumValidation(t *testing.T) {
	tests := []struct {
		name string
		cli  interface{}
		fail bool
	}{
		{
			"Arg",
			&struct {
				Enum string `arg:"" enum:"one,two"`
			}{},
			false,
		},
		{
			"RequiredArg",
			&struct {
				Enum string `required:"" arg:"" enum:"one,two"`
			}{},
			false,
		},
		{
			"OptionalArg",
			&struct {
				Enum string `optional:"" arg:"" enum:"one,two"`
			}{},
			true,
		},
		{
			"RepeatedArgs",
			&struct {
				Enum []string `arg:"" enum:"one,two"`
			}{},
			false,
		},
		{
			"RequiredRepeatedArgs",
			&struct {
				Enum []string `required:"" arg:"" enum:"one,two"`
			}{},
			false,
		},
		{
			"OptionalRepeatedArgs",
			&struct {
				Enum []string `optional:"" arg:"" enum:"one,two"`
			}{},
			false,
		},
		{
			"EnumWithEmptyDefault",
			&struct {
				Flag string `enum:"one,two," default:""`
			}{},
			false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			_, err := kong.New(test.cli)
			if test.fail {
				assert.Error(t, err, repr.String(test.cli))
			} else {
				assert.NoError(t, err, repr.String(test.cli))
			}
		})
	}
}

func TestPassthroughCmd(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		flag    string
		cmdArgs []string
	}{
		{
			"Simple",
			[]string{"--flag", "foobar", "command", "something"},
			"foobar",
			[]string{"something"},
		},
		{
			"DashDash",
			[]string{"--flag", "foobar", "command", "--", "something"},
			"foobar",
			[]string{"--", "something"},
		},
		{
			"Flag",
			[]string{"command", "--flag", "foobar"},
			"",
			[]string{"--flag", "foobar"},
		},
		{
			"FlagAndFlag",
			[]string{"--flag", "foobar", "command", "--flag", "foobar"},
			"foobar",
			[]string{"--flag", "foobar"},
		},
		{
			"NoArgs",
			[]string{"--flag", "foobar", "command"},
			"foobar",
			[]string(nil),
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var cli struct {
				Flag    string
				Command struct {
					Args []string `arg:"" optional:""`
				} `cmd:"" passthrough:""`
			}
			p := mustNew(t, &cli)
			_, err := p.Parse(test.args)
			assert.NoError(t, err)
			assert.Equal(t, test.flag, cli.Flag)
			assert.Equal(t, test.cmdArgs, cli.Command.Args)
		})
	}
}

func TestPassthroughCmdOnlyArgs(t *testing.T) {
	var cli struct {
		Command struct {
			Flag string
			Args []string `arg:"" optional:""`
		} `cmd:"" passthrough:""`
	}
	_, err := kong.New(&cli)
	assert.EqualError(t, err, "<anonymous struct>.Command: passthrough command command [<args> ...] must not have subcommands or flags")
}

func TestPassthroughCmdOnlyStringArgs(t *testing.T) {
	var cli struct {
		Command struct {
			Args []int `arg:"" optional:""`
		} `cmd:"" passthrough:""`
	}
	_, err := kong.New(&cli)
	assert.EqualError(t, err, "<anonymous struct>.Command: passthrough command command [<args> ...] must contain exactly one positional argument of []string type")
}

func TestHelpShouldStillWork(t *testing.T) {
	type CLI struct {
		Dir  string `type:"existingdir" default:"missing-dir"`
		File string `type:"existingfile" default:"testdata/missing.txt"`
	}
	var cli CLI
	w := &strings.Builder{}
	k := mustNew(t, &cli, kong.Writers(w, w))
	rc := -1 // init nonzero to help assert help hook was called
	k.Exit = func(i int) {
		rc = i
	}
	_, err := k.Parse([]string{"--help"})
	t.Log(w.String())
	// checking return code validates the help hook was called
	assert.Zero(t, rc)
	// allow for error propagation from other validation (only for the
	// sake of this test, due to the exit function not actually exiting the
	// program; errors will not propagate in the real world).
	assert.Error(t, err)
}

func TestVersionFlagShouldStillWork(t *testing.T) {
	type CLI struct {
		Dir     string `type:"existingdir" default:"missing-dir"`
		File    string `type:"existingfile" default:"testdata/missing.txt"`
		Version kong.VersionFlag
	}
	var cli CLI
	w := &strings.Builder{}
	k := mustNew(t, &cli, kong.Writers(w, w))
	rc := -1 // init nonzero to help assert help hook was called
	k.Exit = func(i int) {
		rc = i
	}
	_, err := k.Parse([]string{"--version"})
	t.Log(w.String())
	// checking return code validates the help hook was called
	assert.Zero(t, rc)
	// allow for error propagation from other validation (only for the
	// sake of this test, due to the exit function not actually exiting the
	// program; errors will not propagate in the real world).
	assert.Error(t, err)
}

func TestSliceDecoderHelpfulErrorMsg(t *testing.T) {
	tests := []struct {
		name string
		cli  interface{}
		args []string
		err  string
	}{
		{
			"DefaultRune",
			&struct {
				Stuff []string
			}{},
			[]string{"--stuff"},
			`--stuff: missing value, expecting "<arg>,..."`,
		},
		{
			"SpecifiedRune",
			&struct {
				Stuff []string `sep:","`
			}{},
			[]string{"--stuff"},
			`--stuff: missing value, expecting "<arg>,..."`,
		},
		{
			"SpaceRune",
			&struct {
				Stuff []string `sep:" "`
			}{},
			[]string{"--stuff"},
			`--stuff: missing value, expecting "<arg> ..."`,
		},
		{
			"OtherRune",
			&struct {
				Stuff []string `sep:"_"`
			}{},
			[]string{"--stuff"},
			`--stuff: missing value, expecting "<arg>_..."`,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			p := mustNew(t, test.cli)
			_, err := p.Parse(test.args)
			assert.EqualError(t, err, test.err)
		})
	}
}

func TestMapDecoderHelpfulErrorMsg(t *testing.T) {
	tests := []struct {
		name     string
		cli      interface{}
		args     []string
		expected string
	}{
		{
			"DefaultRune",
			&struct {
				Stuff map[string]int
			}{},
			[]string{"--stuff"},
			`--stuff: missing value, expecting "<key>=<value>;..."`,
		},
		{
			"SpecifiedRune",
			&struct {
				Stuff map[string]int `mapsep:";"`
			}{},
			[]string{"--stuff"},
			`--stuff: missing value, expecting "<key>=<value>;..."`,
		},
		{
			"SpaceRune",
			&struct {
				Stuff map[string]int `mapsep:" "`
			}{},
			[]string{"--stuff"},
			`--stuff: missing value, expecting "<key>=<value> ..."`,
		},
		{
			"OtherRune",
			&struct {
				Stuff map[string]int `mapsep:","`
			}{},
			[]string{"--stuff"},
			`--stuff: missing value, expecting "<key>=<value>,..."`,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			p := mustNew(t, test.cli)
			_, err := p.Parse(test.args)
			assert.EqualError(t, err, test.expected)
		})
	}
}

func TestDuplicateName(t *testing.T) {
	var cli struct {
		DupA struct{} `cmd:"" name:"duplicate"`
		DupB struct{} `cmd:"" name:"duplicate"`
	}
	_, err := kong.New(&cli)
	assert.Error(t, err)
}

func TestDuplicateChildName(t *testing.T) {
	var cli struct {
		A struct {
			DupA struct{} `cmd:"" name:"duplicate"`
			DupB struct{} `cmd:"" name:"duplicate"`
		} `cmd:""`
		B struct{} `cmd:""`
	}
	_, err := kong.New(&cli)
	assert.Error(t, err)
}

func TestChildNameCanBeDuplicated(t *testing.T) {
	var cli struct {
		A struct {
			A struct{} `cmd:"" name:"duplicateA"`
			B struct{} `cmd:"" name:"duplicateB"`
		} `cmd:"" name:"duplicateA"`
		B struct{} `cmd:"" name:"duplicateB"`
	}
	mustNew(t, &cli)
}

func TestCumulativeArgumentLast(t *testing.T) {
	var cli struct {
		Arg1 string   `arg:""`
		Arg2 []string `arg:""`
	}
	_, err := kong.New(&cli)
	assert.NoError(t, err)
}

func TestCumulativeArgumentNotLast(t *testing.T) {
	var cli struct {
		Arg2 []string `arg:""`
		Arg1 string   `arg:""`
	}
	_, err := kong.New(&cli)
	assert.Error(t, err)
}

func TestStringPointer(t *testing.T) {
	var cli struct {
		Foo *string
	}
	k, err := kong.New(&cli)
	assert.NoError(t, err)
	assert.NotZero(t, k)
	ctx, err := k.Parse([]string{"--foo", "wtf"})
	assert.NoError(t, err)
	assert.NotZero(t, ctx)
	assert.NotZero(t, cli.Foo)
	assert.Equal(t, "wtf", *cli.Foo)
}

func TestStringPointerNoValue(t *testing.T) {
	var cli struct {
		Foo *string
	}
	k, err := kong.New(&cli)
	assert.NoError(t, err)
	assert.NotZero(t, k)
	ctx, err := k.Parse([]string{})
	assert.NoError(t, err)
	assert.NotZero(t, ctx)
	assert.Zero(t, cli.Foo)
}

func TestStringPointerDefault(t *testing.T) {
	var cli struct {
		Foo *string `default:"stuff"`
	}
	k, err := kong.New(&cli)
	assert.NoError(t, err)
	assert.NotZero(t, k)
	ctx, err := k.Parse([]string{})
	assert.NoError(t, err)
	assert.NotZero(t, ctx)
	assert.NotZero(t, cli.Foo)
	assert.Equal(t, "stuff", *cli.Foo)
}

func TestStringPointerAliasNoValue(t *testing.T) {
	type Foo string
	var cli struct {
		F *Foo
	}
	k, err := kong.New(&cli)
	assert.NoError(t, err)
	assert.NotZero(t, k)
	ctx, err := k.Parse([]string{})
	assert.NoError(t, err)
	assert.NotZero(t, ctx)
	assert.Zero(t, cli.F)
}

func TestStringPointerAlias(t *testing.T) {
	type Foo string
	var cli struct {
		F *Foo
	}
	k, err := kong.New(&cli)
	assert.NoError(t, err)
	assert.NotZero(t, k)
	ctx, err := k.Parse([]string{"--f=value"})
	assert.NoError(t, err)
	assert.NotZero(t, ctx)
	assert.NotZero(t, cli.F)
	assert.Equal(t, Foo("value"), *cli.F)
}

func TestStringPointerEmptyValue(t *testing.T) {
	var cli struct {
		F *string
		G *string
	}
	k, err := kong.New(&cli)
	assert.NoError(t, err)
	assert.NotZero(t, k)
	ctx, err := k.Parse([]string{"--f", "", "--g="})
	assert.NoError(t, err)
	assert.NotZero(t, ctx)
	assert.NotZero(t, cli.F)
	assert.NotZero(t, cli.G)
	assert.Equal(t, "", *cli.F)
	assert.Equal(t, "", *cli.G)
}

func TestIntPtr(t *testing.T) {
	var cli struct {
		F *int
		G *int
	}
	k, err := kong.New(&cli)
	assert.NoError(t, err)
	assert.NotZero(t, k)
	ctx, err := k.Parse([]string{"--f=6"})
	assert.NoError(t, err)
	assert.NotZero(t, ctx)
	assert.NotZero(t, cli.F)
	assert.Zero(t, cli.G)
	assert.Equal(t, 6, *cli.F)
}

func TestBoolPtr(t *testing.T) {
	var cli struct {
		X *bool
	}
	k, err := kong.New(&cli)
	assert.NoError(t, err)
	assert.NotZero(t, k)
	ctx, err := k.Parse([]string{"--x"})
	assert.NoError(t, err)
	assert.NotZero(t, ctx)
	assert.NotZero(t, cli.X)
	assert.Equal(t, true, *cli.X)
}

func TestBoolPtrFalse(t *testing.T) {
	var cli struct {
		X *bool
	}
	k, err := kong.New(&cli)
	assert.NoError(t, err)
	assert.NotZero(t, k)
	ctx, err := k.Parse([]string{"--x=false"})
	assert.NoError(t, err)
	assert.NotZero(t, ctx)
	assert.NotZero(t, cli.X)
	assert.Equal(t, false, *cli.X)
}

func TestBoolPtrNegated(t *testing.T) {
	var cli struct {
		X *bool `negatable:""`
	}
	k, err := kong.New(&cli)
	assert.NoError(t, err)
	assert.NotZero(t, k)
	ctx, err := k.Parse([]string{"--no-x"})
	assert.NoError(t, err)
	assert.NotZero(t, ctx)
	assert.NotZero(t, cli.X)
	assert.Equal(t, false, *cli.X)
}

func TestNilNegatableBoolPtr(t *testing.T) {
	var cli struct {
		X *bool `negatable:""`
	}
	k, err := kong.New(&cli)
	assert.NoError(t, err)
	assert.NotZero(t, k)
	ctx, err := k.Parse([]string{})
	assert.NoError(t, err)
	assert.NotZero(t, ctx)
	assert.Zero(t, cli.X)
}

func TestBoolPtrNil(t *testing.T) {
	var cli struct {
		X *bool
	}
	k, err := kong.New(&cli)
	assert.NoError(t, err)
	assert.NotZero(t, k)
	ctx, err := k.Parse([]string{})
	assert.NoError(t, err)
	assert.NotZero(t, ctx)
	assert.Zero(t, cli.X)
}

func TestUnsupportedPtr(t *testing.T) {
	type Foo struct {
		x int // nolint
		y int // nolint
	}

	var cli struct {
		F *Foo
	}
	k, err := kong.New(&cli)
	assert.NoError(t, err)
	assert.NotZero(t, k)
	ctx, err := k.Parse([]string{"--f=whatever"})
	assert.Zero(t, ctx)
	assert.Error(t, err)
	assert.Equal(t, "--f: cannot find mapper for kong_test.Foo", err.Error())
}

func TestEnumPtr(t *testing.T) {
	var cli struct {
		X *string `enum:"A,B,C" default:"C"`
	}
	k, err := kong.New(&cli)
	assert.NoError(t, err)
	assert.NotZero(t, k)
	ctx, err := k.Parse([]string{"--x=A"})
	assert.NoError(t, err)
	assert.NotZero(t, ctx)
	assert.NotZero(t, cli.X)
	assert.Equal(t, "A", *cli.X)
}

func TestEnumPtrOmitted(t *testing.T) {
	var cli struct {
		X *string `enum:"A,B,C" default:"C"`
	}
	k, err := kong.New(&cli)
	assert.NoError(t, err)
	assert.NotZero(t, k)
	ctx, err := k.Parse([]string{})
	assert.NoError(t, err)
	assert.NotZero(t, ctx)
	assert.NotZero(t, cli.X)
	assert.Equal(t, "C", *cli.X)
}

func TestEnumPtrOmittedNoDefault(t *testing.T) {
	var cli struct {
		X *string `enum:"A,B,C"`
	}
	k, err := kong.New(&cli)
	assert.NoError(t, err)
	assert.NotZero(t, k)
	ctx, err := k.Parse([]string{})
	assert.NoError(t, err)
	assert.NotZero(t, ctx)
	assert.Zero(t, cli.X)
}
