package kong

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestApplyDefaults(t *testing.T) {
	type CLI struct {
		Str      string        `default:"str"`
		Duration time.Duration `default:"30s"`
	}
	cli := &CLI{}
	err := ApplyDefaults(cli)
	require.NoError(t, err)
	require.Equal(t, &CLI{Str: "str", Duration: time.Second * 30}, cli)
}
