package kong_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestIssue556_PositionalArgEnvInHelp(t *testing.T) {
	var cli struct {
		Foo string `arg:"" help:"Foo (${env})" env:"FOO"`
	}
	p := mustNew(t, &cli)
	// Verify ${env} was interpolated to the env var name.
	assert.Equal(t, "Foo (FOO)", p.Model.Positional[0].Help)
}

func TestPositionalArgEnvInHelp_NoEnv(t *testing.T) {
	var cli struct {
		Foo string `arg:"" help:"Foo (${env})"`
	}
	// When no env tag is set, ${env} should interpolate to empty string.
	p := mustNew(t, &cli)
	assert.Equal(t, "Foo ()", p.Model.Positional[0].Help)
}

func TestPositionalArgEnvInHelp_MultipleEnvs(t *testing.T) {
	var cli struct {
		Foo string `arg:"" help:"Foo (${env})" env:"FOO,BAR"`
	}
	// With multiple env vars, ${env} should use the first one.
	p := mustNew(t, &cli)
	assert.Equal(t, "Foo (FOO)", p.Model.Positional[0].Help)
}
