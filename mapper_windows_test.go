//go:build windows
// +build windows

package kong_test

import (
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestWindowsPathMapper(t *testing.T) {
	var cli struct {
		Path  string   `short:"p" type:"path"`
		Files []string `short:"f" type:"path"`
	}
	wd, err := os.Getwd()
	assert.NoError(t, err, "Getwd failed")
	p := mustNew(t, &cli)

	_, err = p.Parse([]string{`-p`, `c:\an\absolute\path`, `-f`, `c:\second\absolute\path\`, `-f`, `relative\path\file`})
	assert.NoError(t, err)
	assert.Equal(t, `c:\an\absolute\path`, cli.Path)
	assert.Equal(t, []string{`c:\second\absolute\path\`, wd + `\relative\path\file`}, cli.Files)
}

func TestWindowsFileMapper(t *testing.T) {
	type CLI struct {
		File *os.File `arg:""`
	}
	var cli CLI
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"testdata\\file.txt"})
	assert.NoError(t, err)
	assert.NotZero(t, cli.File, "File should not be nil")
	_ = cli.File.Close()
	_, err = p.Parse([]string{"testdata\\missing.txt"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing.txt:")
	assert.IsError(t, err, os.ErrNotExist)
	_, err = p.Parse([]string{"-"})
	assert.NoError(t, err)
	assert.Equal(t, os.Stdin, cli.File)
}
