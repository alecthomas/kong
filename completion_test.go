package kong_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dummyCompletionInstaller struct {
	install   func(*kong.Context) error
	uninstall func(*kong.Context) error
}

func (d dummyCompletionInstaller) Install(ctx *kong.Context) error {
	return d.install(ctx)
}

func (d dummyCompletionInstaller) Uninstall(ctx *kong.Context) error {
	return d.uninstall(ctx)
}

func TestCompletionInstallerHelp(t *testing.T) {
	t.Run("visible with defaults", func(t *testing.T) {
		w := &strings.Builder{}
		p := mustNew(t, &struct{}{},
			kong.Writers(w, w),
			kong.Exit(func(int) {}),
			kong.CompletionOptions{
				CompletionInstaller: dummyCompletionInstaller{},
			},
		)
		_, err := p.Parse([]string{"--help"})
		require.NoError(t, err)
		require.Regexp(t, `--install-shell-completion\s+Install shell completion.`, w.String())
		require.Regexp(t, `--uninstall-shell-completion\s+Uninstall shell completion.`, w.String())
	})

	t.Run("visible with overrides", func(t *testing.T) {
		w := &strings.Builder{}
		p := mustNew(t, &struct{}{},
			kong.Writers(w, w),
			kong.Exit(func(int) {}),
			kong.CompletionOptions{
				CompletionInstaller: dummyCompletionInstaller{},
				InstallFlagName:     "override-install-flag",
				InstallFlagHelp:     "override install flag help",
				UninstallFlagName:   "override-uninstall-flag",
				UninstallFlagHelp:   "override uninstall flag help",
			},
		)
		_, err := p.Parse([]string{"--help"})
		require.NoError(t, err)
		require.Regexp(t, `--override-install-flag\s+override install flag help`, w.String())
		require.Regexp(t, `--override-uninstall-flag\s+override uninstall flag help`, w.String())
	})

	t.Run("hidden", func(t *testing.T) {
		w := &strings.Builder{}
		p := mustNew(t, &struct{}{},
			kong.Writers(w, w),
			kong.Exit(func(int) {}),
			kong.ConfigureCompletion(kong.CompletionOptions{
				CompletionInstaller: dummyCompletionInstaller{},
				HideFlags:           true,
			}),
		)
		_, err := p.Parse([]string{"--help"})
		require.NoError(t, err)
		require.NotContains(t, `--install-shell-completion`, w.String())
		require.NotContains(t, `--uninstall-shell-completion`, w.String())
	})

	t.Run("no CompletionsInstaller", func(t *testing.T) {
		w := &strings.Builder{}
		p := mustNew(t, &struct{}{},
			kong.Writers(w, w),
			kong.Exit(func(int) {}),
		)
		_, err := p.Parse([]string{"--help"})
		require.NoError(t, err)
		require.NotContains(t, `--install-shell-completion`, w.String())
		require.NotContains(t, `--uninstall-shell-completion`, w.String())
	})
}

// nolint: dupl
func TestInstallCompletions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var ran, exited bool
		w := &strings.Builder{}
		p := mustNew(t, &struct{}{},
			kong.Writers(w, w),
			kong.Exit(func(i int) {
				exited = assert.Equal(t, 0, i)
			}),
			kong.CompletionOptions{
				CompletionInstaller: dummyCompletionInstaller{
					install: func(context *kong.Context) error {
						ran = true
						return nil
					},
				},
			},
		)
		_, err := p.Parse([]string{"--install-shell-completion"})
		require.NoError(t, err)
		require.True(t, ran)
		require.True(t, exited)
	})

	t.Run("error", func(t *testing.T) {
		var ran bool
		errVal := fmt.Errorf("boo")
		w := &strings.Builder{}
		p := mustNew(t, &struct{}{},
			kong.Writers(w, w),
			kong.CompletionOptions{
				CompletionInstaller: dummyCompletionInstaller{
					install: func(context *kong.Context) error {
						ran = true
						return errVal
					},
				},
			},
		)
		_, err := p.Parse([]string{"--install-shell-completion"})
		require.EqualError(t, err, errVal.Error())
		require.True(t, ran)
	})
}

// nolint: dupl
func TestUninstallCompletions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var ran, exited bool
		w := &strings.Builder{}
		p := mustNew(t, &struct{}{},
			kong.Writers(w, w),
			kong.Exit(func(i int) {
				exited = assert.Equal(t, 0, i)
			}),
			kong.CompletionOptions{
				CompletionInstaller: dummyCompletionInstaller{
					uninstall: func(context *kong.Context) error {
						ran = true
						return nil
					},
				},
			},
		)
		_, err := p.Parse([]string{"--uninstall-shell-completion"})
		require.NoError(t, err)
		require.True(t, ran)
		require.True(t, exited)
	})

	t.Run("error", func(t *testing.T) {
		var ran bool
		errVal := fmt.Errorf("boo")
		w := &strings.Builder{}
		p := mustNew(t, &struct{}{},
			kong.Writers(w, w),
			kong.CompletionOptions{
				CompletionInstaller: dummyCompletionInstaller{
					uninstall: func(context *kong.Context) error {
						ran = true
						return errVal
					},
				},
			},
		)
		_, err := p.Parse([]string{"--uninstall-shell-completion"})
		require.EqualError(t, err, errVal.Error())
		require.True(t, ran)
	})
}

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
