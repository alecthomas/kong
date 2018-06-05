package kong

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestValueMapper(t *testing.T) {
	var cli struct {
		Flag string
	}
	k := mustNew(t, &cli, ValueMapper(&cli.Flag, testMooMapper{}))
	_, err := k.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, "", cli.Flag)
	_, err = k.Parse([]string{"--flag"})
	require.NoError(t, err)
	require.Equal(t, "MOO", cli.Flag)
}

func TestNamedMapper(t *testing.T) {
	var cli struct {
		Flag string `type:"moo"`
	}
	k := mustNew(t, &cli, NamedMapper("moo", testMooMapper{}))
	_, err := k.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, "", cli.Flag)
	_, err = k.Parse([]string{"--flag"})
	require.NoError(t, err)
	require.Equal(t, "MOO", cli.Flag)
}

type testMooMapper struct{}

func (testMooMapper) Decode(ctx *DecoderContext, scan *Scanner, target reflect.Value) error {
	target.SetString("MOO")
	return nil
}
func (testMooMapper) IsBool() bool { return true }

func TestTimeMapper(t *testing.T) {
	var cli struct {
		Flag time.Time `format:"2006"`
	}
	k := mustNew(t, &cli)
	_, err := k.Parse([]string{"--flag=2008"})
	require.NoError(t, err)
	expected, err := time.Parse("2006", "2008")
	require.NoError(t, err)
	require.Equal(t, 2008, expected.Year())
	require.Equal(t, expected, cli.Flag)
}

func TestDurationMapper(t *testing.T) {
	var cli struct {
		Flag time.Duration
	}
	k := mustNew(t, &cli)
	_, err := k.Parse([]string{"--flag=5s"})
	require.NoError(t, err)
	require.Equal(t, time.Second*5, cli.Flag)
}
