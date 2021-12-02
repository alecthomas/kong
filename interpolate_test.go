package kong

import (
	"testing"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	require.Equal(t, `Bobby Brown is 35 years old, 180 cm tall, lives in Melbourne, and likes ${AUD}`, actual)
}

func TestHasInterpolatedVar(t *testing.T) {
	for _, tag := range []string{"name", "age", "height", "city"} {
		require.True(t, HasInterpolatedVar("${name=Bobby Brown} is ${age} years old, ${height} cm tall, lives in ${city=<unknown>}, and likes $${AUD}", tag), tag)
	}

	for _, tag := range []string{"name", "age", "height", "AUD"} {
		require.False(t, HasInterpolatedVar("$name $$age {height} $${AUD}", tag), tag)
	}
}
