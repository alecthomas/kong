//go:build !windows
// +build !windows

package kong_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathMapper(t *testing.T) {
	var cli struct {
		Path string `arg:"" type:"path"`
	}
	p := mustNew(t, &cli)

	_, err := p.Parse([]string{"/an/absolute/path"})
	require.NoError(t, err)
	require.Equal(t, "/an/absolute/path", cli.Path)

	_, err = p.Parse([]string{"-"})
	require.NoError(t, err)
	require.Equal(t, "-", cli.Path)
}
