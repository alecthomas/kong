package kong

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestInterpolate(t *testing.T) {
	vars := map[string]string{
		"age":  "35",
		"city": "Melbourne",
	}
	updatedVars := map[string]string{
		"height": "180",
	}
	actual, err := interpolate("${name=Bobby Brown} is ${age} years old, ${height} cm tall, lives in ${city=<unknown>}, and likes $${AUD}", vars, updatedVars)
	assert.NoError(t, err)
	assert.Equal(t, `Bobby Brown is 35 years old, 180 cm tall, lives in Melbourne, and likes ${AUD}`, actual)
}

func TestInterpolateFallback(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		vars     Vars
		expected string
		err      string
	}{
		{name: "EmptyFallbackUndefined", input: "[${name=}]", expected: "[]"},
		{name: "EmptyFallbackDefined", input: "[${name=}]", vars: Vars{"name": "Bobby"}, expected: "[Bobby]"},
		{name: "NoFallbackUndefined", input: "[${name}]", err: "undefined variable ${name}"},
		{name: "FallbackUndefined", input: "[${name=x}]", expected: "[x]"},
		{name: "FallbackDefined", input: "[${name=x}]", vars: Vars{"name": "Bobby"}, expected: "[Bobby]"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := interpolate(test.input, test.vars, nil)
			if test.err != "" {
				assert.EqualError(t, err, test.err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestHasInterpolatedVar(t *testing.T) {
	assert.True(t, HasInterpolatedVar("${name=}", "name"))

	for _, tag := range []string{"name", "age", "height", "city"} {
		assert.True(t, HasInterpolatedVar("${name=Bobby Brown} is ${age} years old, ${height} cm tall, lives in ${city=<unknown>}, and likes $${AUD}", tag), tag)
	}

	for _, tag := range []string{"name", "age", "height", "AUD"} {
		assert.False(t, HasInterpolatedVar("$name $$age {height} $${AUD}", tag), tag)
	}
}
