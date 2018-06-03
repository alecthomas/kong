package kong

import (
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

type envMap map[string]string

func newEnvParser(t *testing.T, cli interface{}, env envMap) (*Kong, func()) {
	for k, v := range env {
		os.Setenv(k, v)
	}

	r, err := EnvResolver("KONG_")
	require.NoError(t, err)

	parser := mustNew(t, cli, Resolver(r))

	unsetEnvs := func() {
		for k := range env {
			os.Unsetenv(k)
		}
	}

	return parser, unsetEnvs
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
	parser, unsetEnvs := newEnvParser(t, &cli, envMap{"KONG_FLAG": "bye"})
	defer unsetEnvs()

	_, err := parser.Parse([]string{"--flag=hello"})
	require.NoError(t, err)
	require.Equal(t, "hello", cli.Flag)
}

func TestEnvResolverOnlyPopulateUsedBranches(t *testing.T) {
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
	parser, unsetEnvs := newEnvParser(t, &cli, envMap{"KONG_INT": "512"})
	defer unsetEnvs()

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
	parser, unsetEnvs := newEnvParser(t, &cli, envMap{"KONG_NUMBERS": "5,2,9"})
	defer unsetEnvs()

	_, err := parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, []int{5, 2, 9}, cli.Slice)
}

func TestJsonResolverBasic(t *testing.T) {
	var cli struct {
		String string
		Slice  []int
	}

	json := `{
		"string": "🍕",
		"slice": "5,8"
	}`

	r, err := JSONResolver(strings.NewReader(json))
	require.NoError(t, err)

	parser := mustNew(t, &cli, Resolver(r))
	_, err = parser.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "🍕", cli.String)
	require.Equal(t, []int{5, 8}, cli.Slice)
}

//func TestResolversWithHooks(t *testing.T) {
//	require.True(t, false)
//}
//
//func TestResolversWithMappers(t *testing.T) {
//	require.True(t, false)
//}
