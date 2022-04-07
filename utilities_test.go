// Package kong
package kong

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValueOrDefaultValue(t *testing.T) {
	sv := ""
	sd := "Default value"

	// STRING

	s1 := valueOrDefaultValue(sv, sd).(string)
	require.NotEqualValues(t, sv, s1)
	require.EqualValues(t, sd, s1)

	sv = "123"
	s2 := valueOrDefaultValue(sv, sd).(string)
	require.NotEqualValues(t, sd, s2)
	require.EqualValues(t, sv, s2)

	// INT

	iv1 := 0
	id := 715
	i1 := valueOrDefaultValue(iv1, id).(int)
	require.NotEqualValues(t, iv1, i1)
	require.EqualValues(t, id, i1)

	iv1 = 300
	i2 := valueOrDefaultValue(iv1, id).(int)
	require.NotEqualValues(t, id, i2)
	require.EqualValues(t, iv1, i2)

	// BOOL

	bv1 := false
	bd := true
	b1 := valueOrDefaultValue(bv1, bd).(bool)
	require.NotEqualValues(t, bv1, b1)
	require.EqualValues(t, bd, b1)

	bv1, bd = true, false
	b2 := valueOrDefaultValue(bv1, bd).(bool)
	require.NotEqualValues(t, bd, b2)
	require.EqualValues(t, bv1, b2)
}
