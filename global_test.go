package kong

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseHandlingBadBuild(t *testing.T) {
	var cli struct {
		Enabled bool `kong:"fail='"`
	}

	args := os.Args
	defer func() {
		os.Args = args
	}()

	os.Args = []string{os.Args[0], "hi"}

	defer func() {
		if r := recover(); r != nil {
			require.Equal(t, "fail=' is not quoted properly", r.(error).Error())
		}
	}()

	Parse(&cli, Exit(func(_ int) { panic("exiting") }))

	require.Fail(t, "we were expecting a panic")
}
