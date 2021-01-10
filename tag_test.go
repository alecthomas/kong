package kong_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/alecthomas/kong"
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
	_, err := kong.New(&cli)
	require.Error(t, err)
}

func TestNoQuoteEnd(t *testing.T) {
	var cli struct {
		Numbers string `kong:"default='yay"`
	}
	_, err := kong.New(&cli)
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
	// nolint: govet
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
	// nolint: govet
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
	// nolint: govet
	var cli struct {
		Arg string `arg    optional    default:"hi"`
	}

	p := mustNew(t, &cli)
	_, err := p.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "hi", cli.Arg)
}

func TestTagSetOnEmbeddedStruct(t *testing.T) {
	type Embedded struct {
		Key string `help:"A key from ${where}."`
	}
	var cli struct {
		Embedded `set:"where=somewhere"`
	}
	buf := &strings.Builder{}
	p := mustNew(t, &cli, kong.Writers(buf, buf), kong.Exit(func(int) {}))
	_, err := p.Parse([]string{"--help"})
	require.NoError(t, err)
	require.Contains(t, buf.String(), `A key from somewhere.`)
}

func TestTagSetOnCommand(t *testing.T) {
	type Command struct {
		Key string `help:"A key from ${where}."`
	}
	var cli struct {
		Command Command `set:"where=somewhere" cmd:""`
	}
	buf := &strings.Builder{}
	p := mustNew(t, &cli, kong.Writers(buf, buf), kong.Exit(func(int) {}))
	_, err := p.Parse([]string{"command", "--help"})
	require.NoError(t, err)
	require.Contains(t, buf.String(), `A key from somewhere.`)
}

func TestTagSetOnFlag(t *testing.T) {
	var cli struct {
		Flag string `set:"where=somewhere" help:"A key from ${where}."`
	}
	buf := &strings.Builder{}
	p := mustNew(t, &cli, kong.Writers(buf, buf), kong.Exit(func(int) {}))
	_, err := p.Parse([]string{"--help"})
	require.NoError(t, err)
	require.Contains(t, buf.String(), `A key from somewhere.`)
}

func TestTagAliases(t *testing.T) {
	type Command struct {
		Arg string `arg help:"Some arg"`
	}
	var cli struct {
		Cmd Command `cmd aliases:"alias1, alias2"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"alias1", "arg"})
	require.NoError(t, err)
	require.Equal(t, "arg", cli.Cmd.Arg)
	_, err = p.Parse([]string{"alias2", "arg"})
	require.NoError(t, err)
	require.Equal(t, "arg", cli.Cmd.Arg)
}

func TestTagAliasesConflict(t *testing.T) {
	type Command struct {
		Arg string `arg help:"Some arg"`
	}
	var cli struct {
		Cmd      Command `cmd hidden aliases:"other-cmd"`
		OtherCmd Command `cmd`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"other-cmd", "arg"})
	require.NoError(t, err)
	require.Equal(t, "arg", cli.OtherCmd.Arg)
}

func TestTagAliasesSub(t *testing.T) {
	type SubCommand struct {
		Arg string `arg help:"Some arg"`
	}
	type Command struct {
		SubCmd SubCommand `cmd aliases:"other-sub-cmd"`
	}
	var cli struct {
		Cmd Command `cmd hidden`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"cmd", "other-sub-cmd", "arg"})
	require.NoError(t, err)
	require.Equal(t, "arg", cli.Cmd.SubCmd.Arg)
}
