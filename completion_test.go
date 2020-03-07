package kong_test

import (
	"bytes"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alecthomas/kong"
)

func setLineAndPoint(t *testing.T, line string, point *int) func() {
	pVal := len(line)
	if point != nil {
		pVal = *point
	}
	const (
		envLine  = "COMP_LINE"
		envPoint = "COMP_POINT"
	)
	t.Helper()
	origLine, hasOrigLine := os.LookupEnv(envLine)
	origPoint, hasOrigPoint := os.LookupEnv(envPoint)
	require.NoError(t, os.Setenv(envLine, line))
	require.NoError(t, os.Setenv(envPoint, strconv.Itoa(pVal)))
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

	completers := kong.Completers{
		"things":      kong.CompleteSet("thing1", "thing2"),
		"otherthings": kong.CompleteSet("otherthing1", "otherthing2"),
	}

	var cli struct {
		Foo struct {
			Embedded embed  `kong:"embed"`
			Bar      string `kong:"completer=things"`
			Baz      bool
			Rabbit   struct {
			} `kong:"cmd"`
			Duck struct {
			} `kong:"cmd"`
		} `kong:"cmd"`
		Bar struct {
			Tiger   string `kong:"arg,completer=things"`
			Bear    string `kong:"arg,completer=otherthings"`
			OMG     string `kong:"enum='oh,my,gizzles'"`
			Number  int    `kong:"short=n,enum='1,2,3'"`
			BooFlag bool   `kong:"name=boofl,short=b"`
		} `kong:"cmd"`
	}

	type completeTest struct {
		want  []string
		line  string
		point *int
	}

	lenPtr := func(val string) *int {
		v := len(val)
		return &v
	}

	tests := []completeTest{
		{
			line: "myApp ",
			want: []string{"bar", "foo"},
		},
		{
			line: "myApp foo",
			want: []string{"foo"},
		},
		{
			line: "myApp foo ",
			want: []string{"duck", "rabbit"},
		},
		{
			line: "myApp foo r",
			want: []string{"rabbit"},
		},
		{
			line: "myApp foo -",
			want: []string{"--bar", "--baz", "--help", "--lion"},
		},
		{
			line: "myApp foo --lion ",
			want: []string{},
		},
		{
			line: "myApp foo --baz ",
			want: []string{"duck", "rabbit"},
		},
		{
			line: "myApp foo --baz -",
			want: []string{"--bar", "--baz", "--help", "--lion"},
		},
		{
			line: "myApp foo --bar ",
			want: []string{"thing1", "thing2"},
		},
		{
			line: "myApp bar ",
			want: []string{"thing1", "thing2"},
		},
		{
			line: "myApp bar thing",
			want: []string{"thing1", "thing2"},
		},
		{
			line: "myApp bar thing1 ",
			want: []string{"otherthing1", "otherthing2"},
		},
		{
			line: "myApp bar --omg ",
			want: []string{"gizzles", "my", "oh"},
		},
		{
			line: "myApp bar -",
			want: []string{"--boofl", "--help", "--number", "--omg", "-b", "-n"},
		},
		{
			line: "myApp bar -b ",
			want: []string{"thing1", "thing2"},
		},
		{
			line: "myApp bar -b thing1 -",
			want: []string{"--boofl", "--help", "--number", "--omg", "-b", "-n"},
		},
		{
			line: "myApp bar -b thing1 --omg ",
			want: []string{"gizzles", "my", "oh"},
		},
		{
			line: "myApp bar -b thing1 --omg gizzles ",
			want: []string{"otherthing1", "otherthing2"},
		},
		{
			line: "myApp bar -b thing1 --omg gizzles ",
			want: []string{"otherthing1", "otherthing2"},
		},
		{
			line: "myApp bar -b thing1 --omg gi",
			want: []string{"gizzles"},
		},
		{
			line:  "myApp bar -b thing1 --omg gi",
			want:  []string{"thing1", "thing2"},
			point: lenPtr("myApp bar -b th"),
		},
		{
			line:  "myApp bar -b thing1 --omg gizzles ",
			want:  []string{"thing1", "thing2"},
			point: lenPtr("myApp bar -b th"),
		},
		{
			line:  "myApp bar -b thing1 --omg gizzles ",
			want:  []string{"thing1"},
			point: lenPtr("myApp bar -b thing1"),
		},
		{
			line:  "myApp bar -b thing1 --omg gizzles ",
			want:  []string{"otherthing1", "otherthing2"},
			point: lenPtr("myApp bar -b thing1 "),
		},
		{
			line: "myApp bar --number ",
			want: []string{"1", "2", "3"},
		},
		{
			line: "myApp bar --number=",
			want: []string{"1", "2", "3"},
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
				completers,
			)
			cleanup := setLineAndPoint(t, td.line, td.point)
			defer cleanup()
			_, err := p.Parse([]string{})
			require.Error(t, err)
			require.IsType(t, &kong.ParseError{}, err)
			require.True(t, exited)
			require.Equal(t, "", stdErr.String())
			gotLines := strings.Split(stdOut.String(), "\n")
			sort.Strings(gotLines)
			gotOpts := []string{}
			for _, l := range gotLines {
				if l != "" {
					gotOpts = append(gotOpts, l)
				}
			}
			require.Equal(t, td.want, gotOpts)
		})
	}
}
