package kong_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCompletion(t *testing.T) {
	t.Run("completer returns false, nil", func(t *testing.T) {
		var ran bool
		w := &strings.Builder{}
		p := mustNew(t, &struct{}{},
			kong.Writers(w, w),
			kong.Exit(func(i int) {
				require.Fail(t, "exit should not be called")
			}),
			kong.CompletionOptions{
				Completer: func(ctx *kong.Context) (b bool, err error) {
					ran = true
					return false, nil
				},
			},
		)
		_, err := p.Parse([]string{})
		require.NoError(t, err)
		require.True(t, ran)
	})

	t.Run("completer returns true, nil", func(t *testing.T) {
		var ran, exited bool
		output := "foo"
		w := &strings.Builder{}
		p := mustNew(t, &struct{}{},
			kong.Writers(w, w),
			kong.Exit(func(i int) {
				exited = assert.Equal(t, 0, i)
			}),
			kong.CompletionOptions{
				Completer: func(ctx *kong.Context) (bool, error) {
					_, err := ctx.Stdout.Write([]byte(output))
					require.NoError(t, err)
					ran = true
					return true, nil
				},
			},
		)
		_, err := p.Parse([]string{})
		require.NoError(t, err)
		require.True(t, ran)
		require.True(t, exited)
		require.Equal(t, output, w.String())
	})

	t.Run("completer returns true, error", func(t *testing.T) {
		var ran bool
		errVal := fmt.Errorf("boo")
		w := &strings.Builder{}
		p := mustNew(t, &struct{}{},
			kong.Writers(w, w),
			kong.Exit(func(i int) {
				require.Fail(t, "exit should not be called")
			}),
			kong.CompletionOptions{
				Completer: func(ctx *kong.Context) (b bool, err error) {
					ran = true
					return true, errVal
				},
			},
		)
		_, err := p.Parse([]string{})
		require.EqualError(t, err, errVal.Error())
		require.True(t, ran)
	})
}
