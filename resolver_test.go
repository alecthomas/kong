package kong_test

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/alecthomas/kong"
)

type envMap map[string]string

func tempEnv(env envMap) func() {
	for k, v := range env {
		os.Setenv(k, v)
	}

	return func() {
		for k := range env {
			os.Unsetenv(k)
		}
	}
}

func newEnvParser(t *testing.T, cli interface{}, env envMap, options ...kong.Option) (*kong.Kong, func()) {
	t.Helper()
	restoreEnv := tempEnv(env)
	parser := mustNew(t, cli, options...)
	return parser, restoreEnv
}

func TestEnvarsFlagBasic(t *testing.T) {
	var cli struct {
		String string `env:"KONG_STRING"`
		Slice  []int  `env:"KONG_SLICE"`
		Interp string `env:"${kongInterp}"`
	}
	kongInterpEnv := "KONG_INTERP"
	parser, unsetEnvs := newEnvParser(t, &cli,
		envMap{
			"KONG_STRING": "bye",
			"KONG_SLICE":  "5,2,9",
			"KONG_INTERP": "foo",
		},
		kong.Vars{
			"kongInterp": kongInterpEnv,
		},
	)
	defer unsetEnvs()

	_, err := parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "bye", cli.String)
	require.Equal(t, []int{5, 2, 9}, cli.Slice)
	require.Equal(t, "foo", cli.Interp)
}

func TestEnvarsFlagOverride(t *testing.T) {
	var cli struct {
		Flag string `env:"KONG_FLAG"`
	}
	parser, restoreEnv := newEnvParser(t, &cli, envMap{"KONG_FLAG": "bye"})
	defer restoreEnv()

	_, err := parser.Parse([]string{"--flag=hello"})
	require.NoError(t, err)
	require.Equal(t, "hello", cli.Flag)
}

func TestEnvarsTag(t *testing.T) {
	var cli struct {
		Slice []int `env:"KONG_NUMBERS"`
	}
	parser, restoreEnv := newEnvParser(t, &cli, envMap{"KONG_NUMBERS": "5,2,9"})
	defer restoreEnv()

	_, err := parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, []int{5, 2, 9}, cli.Slice)
}

func TestEnvarsEnvPrefix(t *testing.T) {
	type Anonymous struct {
		Slice []int `env:"NUMBERS"`
	}
	var cli struct {
		Anonymous `envprefix:"KONG_"`
	}
	parser, restoreEnv := newEnvParser(t, &cli, envMap{"KONG_NUMBERS": "1,2,3"})
	defer restoreEnv()

	_, err := parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, cli.Slice)
}

func TestEnvarsNestedEnvPrefix(t *testing.T) {
	type NestedAnonymous struct {
		String string `env:"STRING"`
	}
	type Anonymous struct {
		NestedAnonymous `envprefix:"ANON_"`
	}
	var cli struct {
		Anonymous `envprefix:"KONG_"`
	}
	parser, restoreEnv := newEnvParser(t, &cli, envMap{"KONG_ANON_STRING": "abc"})
	defer restoreEnv()

	_, err := parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "abc", cli.String)
}

func TestEnvarsWithDefault(t *testing.T) {
	var cli struct {
		Flag string `env:"KONG_FLAG" default:"default"`
	}
	parser, restoreEnv := newEnvParser(t, &cli, envMap{})
	defer restoreEnv()

	_, err := parser.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, "default", cli.Flag)

	parser, restoreEnv = newEnvParser(t, &cli, envMap{"KONG_FLAG": "moo"})
	defer restoreEnv()
	_, err = parser.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, "moo", cli.Flag)
}

func TestEnv(t *testing.T) {
	type Embed struct {
		Flag string
	}
	type Cli struct {
		One   Embed `prefix:"one-" embed:""`
		Two   Embed `prefix:"two." embed:""`
		Three Embed `prefix:"three_" embed:""`
		Four  Embed `prefix:"four_" embed:""`
		Five  bool
		Six   bool `env:"-"`
	}

	var cli Cli

	expected := Cli{
		One:   Embed{Flag: "one"},
		Two:   Embed{Flag: "two"},
		Three: Embed{Flag: "three"},
		Four:  Embed{Flag: "four"},
		Five:  true,
	}

	// With the prefix
	parser, unsetEnvs := newEnvParser(t, &cli, envMap{
		"KONG_ONE_FLAG":   "one",
		"KONG_TWO_FLAG":   "two",
		"KONG_THREE_FLAG": "three",
		"KONG_FOUR_FLAG":  "four",
		"KONG_FIVE":       "true",
		"KONG_SIX":        "true",
	}, kong.DefaultEnvars("KONG"))
	defer unsetEnvs()

	_, err := parser.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, expected, cli)

	// Without the prefix
	parser, unsetEnvs = newEnvParser(t, &cli, envMap{
		"ONE_FLAG":   "one",
		"TWO_FLAG":   "two",
		"THREE_FLAG": "three",
		"FOUR_FLAG":  "four",
		"FIVE":       "true",
		"SIX":        "true",
	}, kong.DefaultEnvars(""))
	defer unsetEnvs()

	_, err = parser.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, expected, cli)
}

func TestJSONBasic(t *testing.T) {
	type Embed struct {
		String string
	}

	var cli struct {
		String          string
		Slice           []int
		SliceWithCommas []string
		Bool            bool

		One Embed `prefix:"one." embed:""`
		Two Embed `prefix:"two." embed:""`
	}

	json := `{
		"string": "🍕",
		"slice": [5, 8],
		"bool": true,
		"slice_with_commas": ["a,b", "c"],
		"one":{
			"string": "one value"
		},
		"two.string": "two value"
	}`

	r, err := kong.JSON(strings.NewReader(json))
	require.NoError(t, err)

	parser := mustNew(t, &cli, kong.Resolvers(r))
	_, err = parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "🍕", cli.String)
	require.Equal(t, []int{5, 8}, cli.Slice)
	require.Equal(t, []string{"a,b", "c"}, cli.SliceWithCommas)
	require.Equal(t, "one value", cli.One.String)
	require.Equal(t, "two value", cli.Two.String)
	require.True(t, cli.Bool)
}

type testUppercaseMapper struct{}

func (testUppercaseMapper) Decode(ctx *kong.DecodeContext, target reflect.Value) error {
	var value string
	err := ctx.Scan.PopValueInto("lowercase", &value)
	if err != nil {
		return err
	}
	target.SetString(strings.ToUpper(value))
	return nil
}

func TestResolversWithMappers(t *testing.T) {
	var cli struct {
		Flag string `env:"KONG_MOO" type:"upper"`
	}

	restoreEnv := tempEnv(envMap{"KONG_MOO": "meow"})
	defer restoreEnv()

	parser := mustNew(t, &cli,
		kong.NamedMapper("upper", testUppercaseMapper{}),
	)
	_, err := parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "MEOW", cli.Flag)
}

func TestResolverWithBool(t *testing.T) {
	var cli struct {
		Bool bool
	}

	var resolver kong.ResolverFunc = func(context *kong.Context, parent *kong.Path, flag *kong.Flag) (interface{}, error) {
		if flag.Name == "bool" {
			return true, nil
		}
		return nil, nil
	}

	p := mustNew(t, &cli, kong.Resolvers(resolver))

	_, err := p.Parse(nil)
	require.NoError(t, err)
	require.True(t, cli.Bool)
}

func TestLastResolverWins(t *testing.T) {
	var cli struct {
		Int []int
	}

	var first kong.ResolverFunc = func(context *kong.Context, parent *kong.Path, flag *kong.Flag) (interface{}, error) {
		if flag.Name == "int" {
			return 1, nil
		}
		return nil, nil
	}

	var second kong.ResolverFunc = func(context *kong.Context, parent *kong.Path, flag *kong.Flag) (interface{}, error) {
		if flag.Name == "int" {
			return 2, nil
		}
		return nil, nil
	}

	p := mustNew(t, &cli, kong.Resolvers(first, second))
	_, err := p.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, []int{2}, cli.Int)
}

func TestResolverSatisfiesRequired(t *testing.T) {
	var cli struct {
		Int int `required`
	}
	var resolver kong.ResolverFunc = func(context *kong.Context, parent *kong.Path, flag *kong.Flag) (interface{}, error) {
		if flag.Name == "int" {
			return 1, nil
		}
		return nil, nil
	}
	_, err := mustNew(t, &cli, kong.Resolvers(resolver)).Parse(nil)
	require.NoError(t, err)
	require.Equal(t, 1, cli.Int)
}

func TestResolverTriggersHooks(t *testing.T) {
	ctx := &hookContext{}

	var cli struct {
		Flag hookValue
	}

	var first kong.ResolverFunc = func(context *kong.Context, parent *kong.Path, flag *kong.Flag) (interface{}, error) {
		if flag.Name == "flag" {
			return "one", nil
		}
		return nil, nil
	}

	_, err := mustNew(t, &cli, kong.Bind(ctx), kong.Resolvers(first)).Parse(nil)
	require.NoError(t, err)

	require.Equal(t, "one", string(cli.Flag))
	require.Equal(t, []string{"before:", "after:one"}, ctx.values)
}

type validatingResolver struct {
	err error
}

func (v *validatingResolver) Validate(app *kong.Application) error { return v.err }
func (v *validatingResolver) Resolve(context *kong.Context, parent *kong.Path, flag *kong.Flag) (interface{}, error) {
	return nil, nil
}

func TestValidatingResolverErrors(t *testing.T) {
	resolver := &validatingResolver{err: errors.New("invalid")}
	var cli struct{}
	_, err := mustNew(t, &cli, kong.Resolvers(resolver)).Parse(nil)
	require.EqualError(t, err, "invalid")
}
