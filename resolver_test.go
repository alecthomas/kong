package kong

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestEnvResolver(t *testing.T) {
	var cli struct {
		RootFlag string

		Woot struct {
			MySlice []int
			Arg struct {
				Arg string `arg`
				Int int    `env:KONG_TEST`
			} `arg`
		} `cmd`
	}
	parser := mustNew(t, &cli, Resolver(EnvResolver("KONG_")))

	os.Setenv("KONG_ROOT_FLAG", "bye")
	os.Setenv("KONG_MY_SLICE", "9,8")
	os.Setenv("KONG_TEST", "5")

	//err := parser.ApplyResolver(&cli, EnvResolver("KONG_"))
	_, err := parser.Parse([]string{"woot", "arg", "--root-flag=hello"})
	require.NoError(t, err)
	require.Equal(t, "hello", cli.RootFlag)
	require.Equal(t, []int{9, 8}, cli.Woot.MySlice)
	require.Equal(t, 5, cli.Woot.Arg.Int)

	os.Unsetenv("KONG_ROOT_FLAG")
	os.Unsetenv("KONG_MY_SLICE")
	os.Unsetenv("KONG_TEST")
}
