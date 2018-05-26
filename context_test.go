package kong

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTraceErrorPartiallySucceeds(t *testing.T) {
	var cli struct {
		One struct {
			Two struct {
			} `kong:"cmd"`
		} `kong:"cmd"`
	}
	p := mustNew(t, &cli)
	trace, err := Trace([]string{"one", "bad"}, p.Model)
	require.NoError(t, err)
	require.Error(t, trace.Error)
}
