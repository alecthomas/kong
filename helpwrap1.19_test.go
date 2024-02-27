// Wrapping of text changed in Go1.19 per https://github.com/alecthomas/kong/issues/325
// The test has been split pre-go1.19 and go1.19 and onwards.

//go:build go1.19
// +build go1.19

package kong_test

import (
	"bytes"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/kong"
)

func TestCustomWrap(t *testing.T) {
	var cli struct {
		Flag string `help:"A string flag with very long help that wraps a lot and is verbose and is really verbose."`
	}

	w := bytes.NewBuffer(nil)
	app := mustNew(t, &cli,
		kong.Name("test-app"),
		kong.Description("A test app."),
		kong.HelpOptions{
			WrapUpperBound: 50,
		},
		kong.Writers(w, w),
		kong.Exit(func(int) {}),
	)

	_, err := app.Parse([]string{"--help"})
	assert.NoError(t, err)
	expected := `Usage: test-app [flags]

A test app.

Flags:
  -h, --help           Show context-sensitive
                       help.
      --flag=STRING    A string flag with very
                       long help that wraps a
                       lot and is verbose and is
                       really verbose.
`
	t.Log(w.String())
	t.Log(expected)
	assert.Equal(t, expected, w.String())
}
