package kong_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModelApplicationCommands(t *testing.T) {
	var cli struct {
		One struct {
			Two struct {
			} `kong:"cmd"`
			Three struct {
				Four struct {
					Four string `kong:"arg"`
				} `kong:"arg"`
			} `kong:"cmd"`
		} `kong:"cmd"`
	}
	p := mustNew(t, &cli)
	actual := []string{}
	for _, cmd := range p.Model.Leaves() {
		actual = append(actual, cmd.Path())
	}
	require.Equal(t, []string{"one two", "one three <four>"}, actual)
}
