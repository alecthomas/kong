package kong

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInterpolate(t *testing.T) {
	vars := map[string]string{
		"age": "35",
	}
	updatedVars := map[string]string{
		"height": "180",
	}
	actual, err := interpolate("${name=Bobby Brown} is ${age} years old and ${height} cm tall and likes $${AUD}", vars, updatedVars)
	require.NoError(t, err)
	require.Equal(t, `Bobby Brown is 35 years old and 180 cm tall and likes ${AUD}`, actual)
}
