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

func TestConfigOverridesInvalidDefaultForExistingFile(t *testing.T) {
	var cli struct {
		Path string `json:"path" default:"missing-file" type:"existingfile"`
	}

	existingFile := t.TempDir() + "/configured-file"
	assert.NoError(t, os.WriteFile(existingFile, nil, 0o600))
	config := makeConfig(t, map[string]string{"path": existingFile})

	p := mustNew(t, &cli, kong.Configuration(kong.JSON, config))
	_, err := p.Parse(nil)
	assert.NoError(t, err)
	assert.Equal(t, existingFile, cli.Path)
}

func TestConfigDoesNotMaskInvalidDefaultWithoutOverride(t *testing.T) {
	var cli struct {
		Path string `json:"path" default:"missing-file" type:"existingfile"`
	}

	config := makeConfig(t, map[string]string{"other": "value"})
	p := mustNew(t, &cli, kong.Configuration(kong.JSON, config))
	_, err := p.Parse(nil)
	assert.Error(t, err)
}

func TestRegressionIssue489(t *testing.T) {
	type Level struct {
		Flag      string `json:"flag" required:""`
		WithSnake string `json:"with_snake,omitempty" required:""`
		WithCamel string `json:"withCamel,omitempty" required:""`
	}
	var cli struct {
		TopSnake string `json:"top_snake" required:""`
		TopCamel string `json:"topCamel" required:""`
		Level    `json:"level" prefix:"level." embed:""`
	}

	cli.TopCamel = "filled"
	cli.TopSnake = "filled"
	cli.Level.WithCamel = "filled"
	cli.Level.WithSnake = "filled"
	cli.Level.Flag = "filled"
	conf := makeConfig(t, &cli)

	p := mustNew(t, &cli, kong.Configuration(kong.JSON, conf))
	_, err := p.Parse(nil)
	assert.NoError(t, err)
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
