package kong

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelp(t *testing.T) {
	var cli struct {
		String string `kong:"help='A string flag.'"`
		Bool   bool   `kong:"help='A bool flag.'"`

		One struct {
		} `kong:"cmd"`
	}
	w := bytes.NewBuffer(nil)
	exited := false
	app := mustNew(t, &cli,
		Name("test-app"),
		Description("A test app."),
		Writers(w, w),
		ExitFunction(func(int) { exited = true }),
	)
	_, err := app.Parse([]string{"--help"})
	require.NoError(t, err)
	require.True(t, exited)
	fmt.Println(w.String())
}
