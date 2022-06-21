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

func TestHasInterpolatedVar(t *testing.T) {
	for _, tag := range []string{"name", "age", "height", "city"} {
		assert.True(t, HasInterpolatedVar("${name=Bobby Brown} is ${age} years old, ${height} cm tall, lives in ${city=<unknown>}, and likes $${AUD}", tag), tag)
	}

	for _, tag := range []string{"name", "age", "height", "AUD"} {
		assert.False(t, HasInterpolatedVar("$name $$age {height} $${AUD}", tag), tag)
	}
}
