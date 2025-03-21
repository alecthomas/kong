package kong_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/kong"
)

func TestMultipleConfigLoading(t *testing.T) {
	var cli struct {
		Flag string `json:"flag,omitempty"`
	}

	cli.Flag = "first"
	first := makeConfig(t, &cli)

	cli.Flag = ""
	second := makeConfig(t, &cli)

	p := mustNew(t, &cli, kong.Configuration(kong.JSON, first, second))
	_, err := p.Parse(nil)
	assert.NoError(t, err)
	assert.Equal(t, "first", cli.Flag)
}

func TestConfigValidation(t *testing.T) {
	var cli struct {
		Flag string `json:"flag,omitempty" enum:"valid" required:""`
	}

	cli.Flag = "invalid"
	conf := makeConfig(t, &cli)

	p := mustNew(t, &cli, kong.Configuration(kong.JSON, conf))
	_, err := p.Parse(nil)
	assert.Error(t, err)
}

func makeConfig(t *testing.T, config any) (path string) {
	t.Helper()
	w, err := os.CreateTemp(t.TempDir(), "")
	assert.NoError(t, err)
	defer w.Close()
	err = json.NewEncoder(w).Encode(config)
	assert.NoError(t, err)
	return w.Name()
}
