package kong

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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

func newEnvParser(t *testing.T, cli interface{}, env envMap) (*Kong, func()) {
	t.Helper()
	restoreEnv := tempEnv(env)
	parser := mustNew(t, cli, Resolver(PerFlagEnvResolver("KONG_")))
	return parser, restoreEnv
}

func TestEnvResolverFlagBasic(t *testing.T) {
	var cli struct {
		String string
		Slice  []int
	}
	parser, unsetEnvs := newEnvParser(t, &cli, envMap{
		"KONG_STRING": "bye",
		"KONG_SLICE":  "5,2,9",
	})
	defer unsetEnvs()

	_, err := parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "bye", cli.String)
	require.Equal(t, []int{5, 2, 9}, cli.Slice)
}

func TestEnvResolverFlagOverride(t *testing.T) {
	var cli struct {
		Flag string
	}
	parser, restoreEnv := newEnvParser(t, &cli, envMap{"KONG_FLAG": "bye"})
	defer restoreEnv()

	_, err := parser.Parse([]string{"--flag=hello"})
	require.NoError(t, err)
	require.Equal(t, "hello", cli.Flag)
}

func TestEnvResolverOnlyPopulateUsedBranches(t *testing.T) {
	// nolint
	var cli struct {
		UnvisitedArg struct {
			UnvisitedArg string `arg`
			Int          int
		} `arg`
		UnvisitedCmd struct {
			Int int
		} `cmd`
		Visited struct {
			Int int
		} `cmd`
	}
	parser, restoreEnv := newEnvParser(t, &cli, envMap{"KONG_INT": "512"})
	defer restoreEnv()

	_, err := parser.Parse([]string{"visited"})
	require.NoError(t, err)

	require.Equal(t, 512, cli.Visited.Int)
	require.Equal(t, 0, cli.UnvisitedArg.Int)
	require.Equal(t, 0, cli.UnvisitedCmd.Int)
}

func TestEnvResolverTag(t *testing.T) {
	var cli struct {
		Slice []int `env:"KONG_NUMBERS"`
	}
	parser, restoreEnv := newEnvParser(t, &cli, envMap{"KONG_NUMBERS": "5,2,9"})
	defer restoreEnv()

	_, err := parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, []int{5, 2, 9}, cli.Slice)
}

func TestJSONResolverBasic(t *testing.T) {
	var cli struct {
		String          string
		Slice           []int
		SliceWithCommas []string
		Bool            bool
	}

	json := `{
		"string": "🍕",
		"slice": [5, 8],
		"bool": true,
		"slice_with_commas": ["a,b", "c"]
	}`

	r, err := JSONResolver(strings.NewReader(json))
	require.NoError(t, err)

	parser := mustNew(t, &cli, Resolver(r))
	_, err = parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "🍕", cli.String)
	require.Equal(t, []int{5, 8}, cli.Slice)
	require.Equal(t, []string{"a,b", "c"}, cli.SliceWithCommas)
	require.True(t, cli.Bool)
}

func TestResolvedValueTriggersHooks(t *testing.T) {
	var cli struct {
		Int int
	}
	resolver := func(context *Context, parent *Path, flag *Flag) (string, error) {
		if flag.Name == "int" {
			return "1", nil
		}
		return "", nil
	}
	hooked := 0
	p := mustNew(t, &cli, Resolver(resolver), Hook(&cli.Int, func(ctx *Context, path *Path) error {
		hooked++
		return nil
	}))
	_, err := p.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, 1, cli.Int)
	require.Equal(t, 1, hooked)

	hooked = 0
	_, err = p.Parse([]string{"--int=2"})
	require.NoError(t, err)
	require.Equal(t, 2, cli.Int)
	require.Equal(t, 2, hooked)
}

type testUppercaseMapper struct{}

func (testUppercaseMapper) Decode(ctx *DecodeContext, target reflect.Value) error {
	value := ctx.Scan.PopValue("lowercase")
	target.SetString(strings.ToUpper(value))
	return nil
}

func TestResolversWithMappers(t *testing.T) {
	var cli struct {
		Flag string `env:"KONG_MOO" type:"upper"`
	}

	restoreEnv := tempEnv(envMap{"KONG_MOO": "meow"})
	defer restoreEnv()

	r := PerFlagEnvResolver("KONG_")

	parser := mustNew(t, &cli,
		NamedMapper("upper", testUppercaseMapper{}),
		Resolver(r),
	)
	_, err := parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "MEOW", cli.Flag)
}

func TestResolverWithBool(t *testing.T) {
	var cli struct {
		Bool bool
	}

	resolver := func(context *Context, parent *Path, flag *Flag) (string, error) {
		if flag.Name == "bool" {
			return "true", nil
		}
		return "", nil
	}

	p := mustNew(t, &cli, Resolver(resolver))

	_, err := p.Parse(nil)
	require.NoError(t, err)
	require.True(t, cli.Bool)
}

func TestLastResolverWins(t *testing.T) {
	var cli struct {
		Int []int
	}

	var first ResolverFunc = func(context *Context, parent *Path, flag *Flag) (string, error) {
		if flag.Name == "int" {
			return "1", nil
		}
		return "", nil
	}

	var second ResolverFunc = func(context *Context, parent *Path, flag *Flag) (string, error) {
		if flag.Name == "int" {
			return "2", nil
		}
		return "", nil
	}

	p := mustNew(t, &cli, Resolver(first), Resolver(second))
	_, err := p.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, []int{2}, cli.Int)
}

func TestResolverSatisfiesRequired(t *testing.T) {
	var cli struct {
		Int int `required`
	}
	resolver := func(context *Context, parent *Path, flag *Flag) (string, error) {
		if flag.Name == "int" {
			return "1", nil
		}
		return "", nil
	}
	_, err := mustNew(t, &cli, Resolver(resolver)).Parse(nil)
	require.NoError(t, err)
	require.Equal(t, 1, cli.Int)
}

func TestEnvResolver(t *testing.T) {
	var cli struct {
		Int int `env:"SOME_ENVAR"`
	}
	restoreEnv := tempEnv(envMap{"SOME_ENVAR": "12"})
	defer restoreEnv()
	_, err := mustNew(t, &cli, Resolver(EnvResolver())).Parse(nil)
	require.NoError(t, err)
	require.Equal(t, 12, cli.Int)
}
