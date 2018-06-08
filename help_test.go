package kong

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelp(t *testing.T) {
	var cli struct {
		String   string `help:"A string flag."`
		Bool     bool   `help:"A bool flag with very long help that wraps a lot and is verbose and is really verbose."`
		Required bool   `required help:"A required flag."`

		One struct {
			Flag string `help:"Nested flag."`
		} `cmd help:"A subcommand."`

		Two struct {
			Flag        string `help:"Nested flag under two."`
			RequiredTwo bool   `required`

			Three struct {
				RequiredThree bool   `required`
				Three         string `arg`
			} `arg help:"Sub-sub-arg."`

			Four struct {
			} `cmd help:"Sub-sub-command."`
		} `cmd help:"Another subcommand."`
	}

	w := bytes.NewBuffer(nil)
	exited := false
	app := mustNew(t, &cli,
		Name("test-app"),
		Description("A test app."),
		Writers(w, w),
		ExitFunction(func(int) {
			exited = true
			panic(true) // Panic to fake "exit".
		}),
	)

	t.Run("Full", func(t *testing.T) {
		require.Panics(t, func() {
			_, err := app.Parse([]string{"--help"})
			require.NoError(t, err)
		})
		require.True(t, exited)
		t.Log(w.String())
		require.Equal(t, `usage: test-app --required

A test app.

Flags:
  --help           Show context-sensitive help.
  --string=STRING  A string flag.
  --bool           A bool flag with very long help that wraps a lot and is
                   verbose and is really verbose.
  --required       A required flag.

Commands:
  one --required
    A subcommand.

  two <three> --required --required-two --required-three
    Sub-sub-arg.

  two four --required --required-two
    Sub-sub-command.
`, w.String())
	})

	t.Run("Selected", func(t *testing.T) {
		exited = false
		w.Truncate(0)
		require.Panics(t, func() {
			_, err := app.Parse([]string{"two", "hello", "--help"})
			require.NoError(t, err)
		})
		require.True(t, exited)
		t.Log(w.String())
		require.Equal(t, `usage: test-app two <three> --required --required-two --required-three

Sub-sub-arg.

Flags:
  --string=STRING   A string flag.
  --bool            A bool flag with very long help that wraps a lot and is
                    verbose and is really verbose.
  --required        A required flag.

  --flag=STRING     Nested flag under two.
  --required-two

  --required-three
`, w.String())
	})
}
