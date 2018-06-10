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
	restoreEnv := tempEnv(env)

	r, err := EnvResolver("KONG_")
	require.NoError(t, err)

	parser := mustNew(t, cli, Resolver(r))

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
		String string
		Slice  []int
		Bool   bool
	}

	json := `{
		"string": "🍕",
		"slice": [5, 6],
		"bool": true
	}`

	r, err := JSONResolver(strings.NewReader(json))
	require.NoError(t, err)

	parser := mustNew(t, &cli, Resolver(r))
	_, err = parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "🍕", cli.String)
	require.Equal(t, []int{5, 8}, cli.Slice)
	require.True(t, cli.Bool)
}

func TestResolversWithHooks(t *testing.T) {
	require.True(t, false)
}

type testUppercaseMapper struct{}

func (testUppercaseMapper) Decode(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
	value := scan.PopValue("lowercase")
	target.SetString(strings.ToUpper(value))
	return nil
}

func TestResolversWithMappers(t *testing.T) {
	var cli struct {
		Flag string `env:"KONG_MOO" type:"upper"`
	}

	restoreEnv := tempEnv(envMap{"KONG_MOO": "meow"})
	defer restoreEnv()

	r, _ := EnvResolver("KONG_")

	parser := mustNew(t, &cli,
		NamedMapper("upper", testUppercaseMapper{}),
		Resolver(r),
	)
	_, err := parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "MEOW", cli.Flag)
}
