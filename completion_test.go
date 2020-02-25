package kong_test

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setLineAndPoint(t *testing.T, line string) func() {
	const (
		envLine  = "COMP_LINE"
		envPoint = "COMP_POINT"
	)
	t.Helper()
	origLine, hasOrigLine := os.LookupEnv(envLine)
	origPoint, hasOrigPoint := os.LookupEnv(envPoint)
	require.NoError(t, os.Setenv(envLine, line))
	require.NoError(t, os.Setenv(envPoint, strconv.Itoa(len(line))))
	return func() {
		t.Helper()
		require.NoError(t, os.Unsetenv(envLine))
		require.NoError(t, os.Unsetenv(envPoint))
		if hasOrigLine {
			require.NoError(t, os.Setenv(envLine, origLine))
		}
		if hasOrigPoint {
			require.NoError(t, os.Setenv(envPoint, origPoint))
		}
	}
}

func TestComplete(t *testing.T) {
	type embed struct {
		Lion string
	}

	predictors := map[string]kong.Predictor{
		"things":      kong.PredictSet("thing1", "thing2"),
		"otherthings": kong.PredictSet("otherthing1", "otherthing2"),
	}

	var cli struct {
		Foo struct {
			Embedded embed  `kong:"embed"`
			Bar      string `kong:"predictor=things"`
			Baz      bool
			Rabbit   struct {
			} `kong:"cmd"`
			Duck struct {
			} `kong:"cmd"`
		} `kong:"cmd"`
		Bar struct {
			Tiger   string `kong:"arg,predictor=things"`
			Bear    string `kong:"arg,predictor=otherthings"`
			OMG     string `kong:"enum='oh,my,gizzles'"`
			Number  int    `kong:"short=n,enum='1,2,3'"`
			BooFlag bool   `kong:"name=boofl,short=b"`
		} `kong:"cmd"`
	}

	type completeTest struct {
		want []string
		line string
	}

	tests := []completeTest{
		{
			want: []string{"foo", "bar"},
			line: "myApp ",
		},
		{
			want: []string{"foo"},
			line: "myApp foo",
		},
		{
			want: []string{"rabbit", "duck"},
			line: "myApp foo ",
		},
		{
			want: []string{"rabbit"},
			line: "myApp foo r",
		},
		{
			want: []string{"--bar", "--baz", "--lion", "--help"},
			line: "myApp foo -",
		},
		{
			want: []string{},
			line: "myApp foo --lion ",
		},
		{
			want: []string{"rabbit", "duck"},
			line: "myApp foo --baz ",
		},
		{
			want: []string{"--bar", "--baz", "--lion", "--help"},
			line: "myApp foo --baz -",
		},
		{
			want: []string{"thing1", "thing2"},
			line: "myApp foo --bar ",
		},
		{
			want: []string{"thing1", "thing2"},
			line: "myApp bar ",
		},
		{
			want: []string{"thing1", "thing2"},
			line: "myApp bar thing",
		},
		{
			want: []string{"otherthing1", "otherthing2"},
			line: "myApp bar thing1 ",
		},
		{
			want: []string{"oh", "my", "gizzles"},
			line: "myApp bar --omg ",
		},
		{
			want: []string{"-n", "--number", "--omg", "--help", "--boofl", "-b"},
			line: "myApp bar -",
		},
		{
			want: []string{"thing1", "thing2"},
			line: "myApp bar -b ",
		},
		{
			want: []string{"-n", "--number", "--omg", "--help", "--boofl", "-b"},
			line: "myApp bar -b thing1 -",
		},
		{
			want: []string{"oh", "my", "gizzles"},
			line: "myApp bar -b thing1 --omg ",
		},
		{
			want: []string{"otherthing1", "otherthing2"},
			line: "myApp bar -b thing1 --omg gizzles ",
		},
	}
	for _, td := range tests {
		td := td
		t.Run(td.line, func(t *testing.T) {
			var stdOut, stdErr bytes.Buffer
			var exited bool
			p := mustNew(t, &cli,
				kong.Writers(&stdOut, &stdErr),
				kong.Exit(func(i int) {
					exited = assert.Equal(t, 0, i)
				}),
				kong.CompletionOptions{
					RunCompletion: true,
					Predictors:    predictors,
				},
			)
			cleanup := setLineAndPoint(t, td.line)
			defer cleanup()
			_, err := p.Parse([]string{})
			require.Error(t, err)
			require.IsType(t, &kong.ParseError{}, err)
			require.True(t, exited)
			require.Equal(t, "", stdErr.String())
			gotLines := strings.Split(stdOut.String(), "\n")
			gotOpts := []string{}
			for _, l := range gotLines {
				if l != "" {
					gotOpts = append(gotOpts, l)
				}
			}
			require.ElementsMatch(t, td.want, gotOpts)
		})
	}
}

func TestRunCompletion(t *testing.T) {
	t.Run("completer returns false, nil", func(t *testing.T) {
		var ran bool
		w := &strings.Builder{}
		p := mustNew(t, &struct{}{},
			kong.Writers(w, w),
			kong.CompletionOptions{
				RunCompletion: true,
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
				RunCompletion: true,
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
			kong.CompletionOptions{
				RunCompletion: true,
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

	t.Run("RunCompletion = false", func(t *testing.T) {
		w := &strings.Builder{}
		p := mustNew(t, &struct{}{},
			kong.Writers(w, w),
			kong.CompletionOptions{
				Completer: func(ctx *kong.Context) (b bool, err error) {
					require.Fail(t, "completer should not be called")
					return true, nil
				},
			},
		)
		_, err := p.Parse([]string{})
		require.NoError(t, err)
	})
}

// nolint: dupl
func TestInstallCompletion(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var ran bool
		var cli struct {
			InstallCompletion kong.InstallCompletion `kong:"cmd"`
		}
		installer := kong.NewCompletionInstaller(func(ctx *kong.Context) error {
			ran = true
			return nil
		}, nil)
		p := mustNew(t, &cli,
			kong.CompletionOptions{
				CompletionInstaller: installer,
			},
		)
		k, err := p.Parse([]string{"install-completion"})
		require.NoError(t, err)
		err = k.Run()
		require.NoError(t, err)
		require.True(t, ran)
	})

	t.Run("error", func(t *testing.T) {
		var ran bool
		errVal := fmt.Errorf("boo")
		var cli struct {
			InstallCompletion kong.InstallCompletion `kong:"cmd"`
		}
		installer := kong.NewCompletionInstaller(func(ctx *kong.Context) error {
			ran = true
			return errVal
		}, nil)
		p := mustNew(t, &cli,
			kong.CompletionOptions{
				CompletionInstaller: installer,
			},
		)
		k, err := p.Parse([]string{"install-completion"})
		require.NoError(t, err)
		err = k.Run()
		assert.EqualError(t, err, errVal.Error())
		require.True(t, ran)
	})
}

// nolint: dupl
func TestUninstallCompletion(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var ran bool
		var cli struct {
			UninstallCompletion kong.UninstallCompletion `kong:"cmd"`
		}
		installer := kong.NewCompletionInstaller(nil, func(ctx *kong.Context) error {
			ran = true
			return nil
		})
		p := mustNew(t, &cli,
			kong.CompletionOptions{
				CompletionInstaller: installer,
			},
		)
		k, err := p.Parse([]string{"uninstall-completion"})
		require.NoError(t, err)
		err = k.Run()
		require.NoError(t, err)
		require.True(t, ran)
	})

	t.Run("error", func(t *testing.T) {
		var ran bool
		errVal := fmt.Errorf("boo")
		var cli struct {
			UninstallCompletion kong.UninstallCompletion `kong:"cmd"`
		}
		installer := kong.NewCompletionInstaller(nil, func(ctx *kong.Context) error {
			ran = true
			return errVal
		})
		p := mustNew(t, &cli,
			kong.CompletionOptions{
				CompletionInstaller: installer,
			},
		)
		k, err := p.Parse([]string{"uninstall-completion"})
		require.NoError(t, err)
		err = k.Run()
		assert.EqualError(t, err, errVal.Error())
		require.True(t, ran)
	})
}
