package kong

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigFlag(t *testing.T) {
	var cli struct {
		Config ConfigFlag
		Flag   string
	}

	w, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(w.Name())
	w.WriteString(`{"flag": "hello world"}`) // nolint: errcheck
	w.Close()

	p := Must(&cli, Configuration(JSON))
	_, err = p.Parse([]string{"--config", w.Name()})
	require.NoError(t, err)
	require.Equal(t, "hello world", cli.Flag)
}

func TestVersionFlag(t *testing.T) {
	var cli struct {
		Version VersionFlag
	}
	w := &strings.Builder{}
	p := Must(&cli, Vars{"version": "0.1.1"})
	p.Stdout = w
	called := 1
	p.Exit = func(s int) { called = s }

	_, err := p.Parse([]string{"--version"})
	require.NoError(t, err)
	require.Equal(t, "0.1.1", strings.TrimSpace(w.String()))
	require.Equal(t, 0, called)
}

func TestChangeDirFlag(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(cwd) // nolint: errcheck

	dir := t.TempDir()
	file := filepath.Join(dir, "out.txt")
	err = os.WriteFile(file, []byte("foobar"), 0o600)
	require.NoError(t, err)

	var cli struct {
		ChangeDir ChangeDirFlag `short:"C"`
		Path      string        `arg:"" type:"existingfile"`
	}

	p := Must(&cli)
	_, err = p.Parse([]string{"-C", dir, "out.txt"})
	require.NoError(t, err)
	require.Equal(t, file, cli.Path)
}
