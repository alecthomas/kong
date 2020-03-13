package kong

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTestFilesDir(t *testing.T) (teardown func()) {
	t.Helper()
	var err error
	tmpDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	files := []string{
		"dir/foo",
		"dir/bar",
		"outer/inner/readme.md",
		".dot.txt",
		"a.txt",
		"b.txt",
		"c.txt",
		"readme.md",
	}
	for _, file := range files {
		file = filepath.Join(tmpDir, filepath.FromSlash(file))
		require.NoError(t, os.MkdirAll(filepath.Dir(file), 0700))
		require.NoError(t, ioutil.WriteFile(file, nil, 0600))
	}
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	return func() {
		require.NoError(t, os.Chdir(wd))
		require.NoError(t, os.RemoveAll(tmpDir))
	}
}

func TestCompleteFilesSet(t *testing.T) {
	set := CompleteFilesSet([]string{
		"foo/bar", "baz", "foo/qux", ".", "./one", "./flood",
	})

	for _, td := range []struct {
		input string
		want  []string
	}{
		{
			input: "f",
			want:  []string{"flood", "foo/bar", "foo/qux"},
		},
		{
			input: "./f",
			want:  []string{"./flood", "./foo/bar", "./foo/qux"},
		},
		{
			input: "foo/",
			want:  []string{"foo/bar", "foo/qux"},
		},
		{
			input: "./foo/",
			want:  []string{"./foo/bar", "./foo/qux"},
		},
		{
			input: "",
			want:  []string{"./", "baz", "flood", "foo/bar", "foo/qux", "one"},
		},
		{
			input: ".",
			want:  []string{"./", "./baz", "./flood", "./foo/bar", "./foo/qux", "./one"},
		},
		{
			input: "./",
			want:  []string{"./", "./baz", "./flood", "./foo/bar", "./foo/qux", "./one"},
		},
		{
			input: "./q",
			want:  []string{},
		},
		{
			input: "q",
			want:  []string{},
		},
		{
			input: "foo/q",
			want:  []string{"foo/qux"},
		},
		{
			input: "./foo/q",
			want:  []string{"./foo/qux"},
		},
	} {
		td := td
		t.Run(fmt.Sprintf("input:%q", td.input), func(t *testing.T) {
			input := filepath.FromSlash(td.input)
			got := set.Options(newCompleterArgs(input))
			sort.Strings(got)
			sort.Strings(td.want)
			require.Equal(t, td.want, got)
		})
	}
}

func TestCompleteDirs(t *testing.T) {
	teardown := setupTestFilesDir(t)
	defer teardown()
	for pattern, args := range map[string]map[string][]string{
		"*": {
			"di":     {"dir/"},
			"dir":    {"dir/"},
			"dir/":   {"dir/"},
			"./di":   {"./dir/"},
			"./dir":  {"./dir/"},
			"./dir/": {"./dir/"},
			"":       {"./", "dir/", "outer/"},
			".":      {"./", "./dir/", "./outer/"},
			"./":     {"./", "./dir/", "./outer/"},
		},
		"*.md": {
			"ou":       {"outer/", "outer/inner/"},
			"outer":    {"outer/", "outer/inner/"},
			"outer/":   {"outer/", "outer/inner/"},
			"./ou":     []string{"./outer/", "./outer/inner/"},
			"./outer":  []string{"./outer/", "./outer/inner/"},
			"./outer/": []string{"./outer/", "./outer/inner/"},
		},
		"dir": {
			"di":     {"dir/"},
			"dir":    {"dir/"},
			"dir/":   {"dir/"},
			"./di":   {"./dir/"},
			"./dir":  {"./dir/"},
			"./dir/": {"./dir/"},
		},
	} {
		pattern := pattern
		args := args
		t.Run(fmt.Sprintf("pattern:%q", pattern), func(t *testing.T) {
			completer := CompleteDirs()
			for arg, want := range args {
				arg := arg
				want := want
				t.Run(fmt.Sprintf("arg:%q", arg), func(t *testing.T) {
					got := completer.Options(newCompleterArgs(arg))
					sort.Strings(got)
					sort.Strings(want)
					require.Equal(t, want, got)
				})
			}
		})
	}
}

func TestCompleteFiles(t *testing.T) {
	teardown := setupTestFilesDir(t)
	defer teardown()
	for pattern, args := range map[string]map[string][]string{
		"*.txt": {
			"":       {"./", "dir/", "outer/", "a.txt", "b.txt", "c.txt", ".dot.txt"},
			"./dir/": []string{"./dir/"},
		},
		"*": {
			"./dir/f":   []string{"./dir/foo"},
			"./dir/foo": []string{"./dir/foo"},
			"dir":       []string{"dir/", "dir/foo", "dir/bar"},
			"di":        []string{"dir/", "dir/foo", "dir/bar"},
			"dir/":      []string{"dir/", "dir/foo", "dir/bar"},
			"./dir":     []string{"./dir/", "./dir/foo", "./dir/bar"},
			"./dir/":    []string{"./dir/", "./dir/foo", "./dir/bar"},
			"./di":      []string{"./dir/", "./dir/foo", "./dir/bar"},
		},
		"*.md": {
			"":        []string{"./", "dir/", "outer/", "readme.md"},
			".":       []string{"./", "./dir/", "./outer/", "./readme.md"},
			"./":      []string{"./", "./dir/", "./outer/", "./readme.md"},
			"outer/i": []string{"outer/inner/", "outer/inner/readme.md"},
		},
		"foo": {
			"./dir/": []string{"./dir/", "./dir/foo"},
			"./d":    []string{"./dir/", "./dir/foo"},
		},
	} {
		pattern := pattern
		args := args
		t.Run(fmt.Sprintf("pattern:%q", pattern), func(t *testing.T) {
			completer := CompleteFiles(pattern)
			for arg, want := range args {
				arg := arg
				want := want
				t.Run(fmt.Sprintf("arg:%q", arg), func(t *testing.T) {
					got := completer.Options(newCompleterArgs(arg))
					sort.Strings(got)
					sort.Strings(want)
					require.Equal(t, want, got)
				})
			}
		})
	}
}
