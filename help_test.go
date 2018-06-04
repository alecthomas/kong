package kong

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelp(t *testing.T) {
	var cli struct {
		String string `help:"A string flag."`
		Bool   bool   `help:"A bool flag with very long help that wraps a lot and is verbose and is really verbose."`

		One struct {
			Flag string `help:"Nested flag."`
		} `cmd help:"A subcommand."`

		Two struct {
			Flag string `help:"Nested flag under two."`
		} `cmd help:"Another subcommand."`
	}
	w := bytes.NewBuffer(nil)
	exited := false
	app := mustNew(t, &cli,
		Name("test-app"),
		Description("A test app."),
		Writers(w, w),
		ExitFunction(func(int) { exited = true }),
	)
	_, err := app.Parse([]string{"--help"})
	require.NoError(t, err)
	require.True(t, exited)
	require.Equal(t, `usage: test-app [<flags>]

A test app.

Flags:
  --help           Show context-sensitive help.
  --string=STRING  A string flag.
  --bool           A bool flag with very long help that wraps a lot and is
                   verbose and is really verbose.

Commands:
  one [<flags>]
    A subcommand.

  two [<flags>]
    Another subcommand.
`, w.String())

	exited = false
	w.Truncate(0)
	_, err = app.Parse([]string{"one", "--help"})
	require.NoError(t, err)
	require.True(t, exited)
	require.Equal(t, `usage: test-app one [<flags>]

A subcommand.

Flags:
  --flag=STRING  Nested flag.
`, w.String())
}
