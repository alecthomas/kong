package kong

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInterpolate(t *testing.T) {
	vars := map[string]string{
		"age": "35",
	}
	actual, err := interpolate("${name=Bobby Brown} is ${age} years old", vars)
	require.NoError(t, err)
	require.Equal(t, `Bobby Brown is 35 years old`, actual)
}
