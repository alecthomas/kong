package kong

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestConfigFlag(t *testing.T) {
	var cli struct {
		Config ConfigFlag
		Flag   string
	}

	w, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer os.Remove(w.Name())
	w.WriteString(`{"flag": "hello world"}`) //nolint: errcheck
	w.Close()

	p := Must(&cli, Configuration(JSON))
	_, err = p.Parse([]string{"--config", w.Name()})
	assert.NoError(t, err)
	assert.Equal(t, "hello world", cli.Flag)
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
	assert.NoError(t, err)
	assert.Equal(t, "0.1.1", strings.TrimSpace(w.String()))
	assert.Equal(t, 0, called)
}

func TestChangeDirFlag(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(cwd) //nolint: errcheck

	dir := t.TempDir()
	file := filepath.Join(dir, "out.txt")
	err = os.WriteFile(file, []byte("foobar"), 0o600)
	assert.NoError(t, err)

	var cli struct {
		ChangeDir ChangeDirFlag `short:"C"`
		Path      string        `arg:"" type:"existingfile"`
	}

	p := Must(&cli)
	_, err = p.Parse([]string{"-C", dir, "out.txt"})
	assert.NoError(t, err)
	if runtime.GOOS != "windows" {
		file, err = filepath.EvalSymlinks(file) // Needed because OSX uses a symlinked tmp dir.
		assert.NoError(t, err)
	}
	assert.Equal(t, file, cli.Path)
}
