package kong

import (
	"bytes"
	"compress/gzip"
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

func TestUnzip(t *testing.T) {
	inputData := "hello batman"

	// Compress inputData
	var writeBuffer bytes.Buffer

	writer := gzip.NewWriter(&writeBuffer)
	_, writeErr := writer.Write([]byte(inputData))
	closeErr := writer.Close()
	require.NoError(t, closeErr)
	require.NoError(t, writeErr)
	compressedInputData := writeBuffer.Bytes()

	// Now try to decompress
	unzippedData, unzipErr := Unzip(compressedInputData)
	require.NoError(t, unzipErr)
	require.Equal(t, string(unzippedData), inputData)
}
