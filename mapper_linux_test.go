//go:build !windows
// +build !windows

package kong_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestPathMapper(t *testing.T) {
	var cli struct {
		Path string `arg:"" type:"path"`
	}
	p := mustNew(t, &cli)

	_, err := p.Parse([]string{"/an/absolute/path"})
	assert.NoError(t, err)
	assert.Equal(t, "/an/absolute/path", cli.Path)

	_, err = p.Parse([]string{"-"})
	assert.NoError(t, err)
	assert.Equal(t, "-", cli.Path)
}
