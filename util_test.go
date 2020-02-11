package kong

import (
	"fmt"
	"io/ioutil"
	"os"
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

func TestDebugConfigFlag(t *testing.T) {
	var cli struct {
		DebugConfig DebugConfigFlag
		Flag        string
	}

	w, err := ioutil.TempFile("", "config-")
	require.NoError(t, err)
	defer os.Remove(w.Name())
	p := Must(&cli, Configuration(JSON, w.Name(), "/does/not/exist.yaml"))
	stderr := &strings.Builder{}
	p.Stderr = stderr
	w.WriteString(`{"flag": "hello world"}`) // nolint: errcheck
	w.Close()

	_, err = p.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "", stderr.String())

	_, err = p.Parse([]string{"--debug-config"})
	require.NoError(t, err)
	expectedOut := fmt.Sprintf("Parsed configuration files:\n%s%s\n", SpaceIndenter(""), w.Name())
	require.Equal(t, expectedOut, stderr.String())
}
