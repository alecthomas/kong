package kong

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultValueForOptionalArg(t *testing.T) {
	var cli struct {
		Arg string `kong:"arg,optional,default='👌'"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, "👌", cli.Arg)
}

func TestNoValueInTag(t *testing.T) {
	var cli struct {
		Empty1 string `kong:"default"`
		Empty2 string `kong:"default="`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, "", cli.Empty1)
	require.Equal(t, "", cli.Empty2)
}

func TestCommaInQuotes(t *testing.T) {
	var cli struct {
		Numbers string `kong:"default='1,2'"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, "1,2", cli.Numbers)
}

func TestBadString(t *testing.T) {
	var cli struct {
		Numbers string `kong:"default='yay'n"`
	}
	_, err := New(&cli)
	require.Error(t, err)
}

func TestNoQuoteEnd(t *testing.T) {
	var cli struct {
		Numbers string `kong:"default='yay"`
	}
	_, err := New(&cli)
	require.Error(t, err)
}

func TestEscapedQuote(t *testing.T) {
	var cli struct {
		DoYouKnow string `kong:"default='i don\\'t know'"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, "i don't know", cli.DoYouKnow)
}

func TestBareTags(t *testing.T) {
	var cli struct {
		Cmd struct {
			Arg  string `arg`
			Flag string `required default:"👌"`
		} `cmd`
	}

	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"cmd", "arg", "--flag=hi"})
	require.NoError(t, err)
	require.Equal(t, "hi", cli.Cmd.Flag)
	require.Equal(t, "arg", cli.Cmd.Arg)
}

func TestBareTagsWithJsonTag(t *testing.T) {
	var cli struct {
		Cmd struct {
			Arg  string `json:"-" optional arg`
			Flag string `json:"best_flag" default:"\"'👌'\""`
		} `cmd json:"CMD"`
	}

	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"cmd"})
	require.NoError(t, err)
	require.Equal(t, "\"'👌'\"", cli.Cmd.Flag)
	require.Equal(t, "", cli.Cmd.Arg)
}

func TestManySeps(t *testing.T) {
	var cli struct {
		Arg string `arg    optional    default:"hi"`
	}

	p := mustNew(t, &cli)
	_, err := p.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "hi", cli.Arg)
}
