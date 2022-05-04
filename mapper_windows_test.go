//go:build windows
// +build windows

package kong_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWindowsPathMapper(t *testing.T) {
	var cli struct {
		Path  string   `short:"p" type:"path"`
		Files []string `short:"f" type:"path"`
	}
	wd, err := os.Getwd()
	require.NoError(t, err, "Getwd failed")
	p := mustNew(t, &cli)

	_, err = p.Parse([]string{`-p`, `c:\an\absolute\path`, `-f`, `c:\second\absolute\path\`, `-f`, `relative\path\file`})
	require.NoError(t, err)
	require.Equal(t, `c:\an\absolute\path`, cli.Path)
	require.Equal(t, []string{`c:\second\absolute\path\`, wd + `\relative\path\file`}, cli.Files)
}

func TestWindowsFileMapper(t *testing.T) {
	type CLI struct {
		File *os.File `arg:""`
	}
	var cli CLI
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"testdata\\file.txt"})
	require.NoError(t, err)
	require.NotNil(t, cli.File)
	_ = cli.File.Close()
	_, err = p.Parse([]string{"testdata\\missing.txt"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing.txt: The system cannot find the file specified.")
	_, err = p.Parse([]string{"-"})
	require.NoError(t, err)
	require.Equal(t, os.Stdin, cli.File)
}
