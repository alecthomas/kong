package kong

import (
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"
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
			assert.Equal(t, "fail=' is not quoted properly", r.(error).Error()) //nolint
		}
	}()

	Parse(&cli, Exit(func(_ int) { panic("exiting") }))

	t.Fatal("we were expecting a panic")
}
