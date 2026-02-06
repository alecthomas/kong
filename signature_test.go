package kong_test

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/alecthomas/kong"
)

// Command with value receiver.
type signatureCmd struct {
	Arg string `arg:"" optional:""`
}

func (signatureCmd) Signature() string {
	return `cmd:"" name:"sig" help:"Signature help" aliases:"s,sg"`
}

func TestSignatureCommand(t *testing.T) {
	var cli struct {
		Cmd signatureCmd
	}
	p := mustNew(t, &cli)

	// Should be reachable by the signature-provided name.
	_, err := p.Parse([]string{"sig", "value"})
	assert.NoError(t, err)
	assert.Equal(t, "value", cli.Cmd.Arg)

	// Should be reachable by aliases.
	_, err = p.Parse([]string{"s", "other"})
	assert.NoError(t, err)
	assert.Equal(t, "other", cli.Cmd.Arg)

	_, err = p.Parse([]string{"sg", "third"})
	assert.NoError(t, err)
	assert.Equal(t, "third", cli.Cmd.Arg)
}

func TestSignatureCommandHelp(t *testing.T) {
	var cli struct {
		Cmd signatureCmd
	}
	buf := &strings.Builder{}
	p := mustNew(t, &cli, kong.Writers(buf, buf), kong.Exit(func(int) {}))
	_, _ = p.Parse([]string{"--help"})
	assert.Contains(t, buf.String(), "Signature help")
}

// Command with pointer receiver.
type signaturePtrCmd struct {
	Flag string `default:"def"`
}

func (*signaturePtrCmd) Signature() string {
	return `cmd:"" name:"ptrcmd" help:"Pointer receiver help"`
}

func TestSignaturePointerReceiver(t *testing.T) {
	var cli struct {
		Cmd *signaturePtrCmd
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"ptrcmd"})
	assert.NoError(t, err)
	assert.Equal(t, "def", cli.Cmd.Flag)
}

func TestSignaturePointerReceiverHelp(t *testing.T) {
	var cli struct {
		Cmd *signaturePtrCmd
	}
	buf := &strings.Builder{}
	p := mustNew(t, &cli, kong.Writers(buf, buf), kong.Exit(func(int) {}))
	_, _ = p.Parse([]string{"--help"})
	assert.Contains(t, buf.String(), "Pointer receiver help")
}

func TestSignatureFieldTagOverrides(t *testing.T) {
	var cli struct {
		// Field tag overrides the name from the signature, but help comes from signature.
		Cmd signatureCmd `cmd:"" name:"override"`
	}
	p := mustNew(t, &cli)

	// The overridden name should work.
	_, err := p.Parse([]string{"override", "val"})
	assert.NoError(t, err)
	assert.Equal(t, "val", cli.Cmd.Arg)

	// The signature name should NOT work because the field tag overrode it.
	_, err = p.Parse([]string{"sig"})
	assert.Error(t, err)
}

func TestSignatureFieldTagOverridesHelp(t *testing.T) {
	var cli struct {
		Cmd signatureCmd `cmd:"" name:"override"`
	}
	buf := &strings.Builder{}
	p := mustNew(t, &cli, kong.Writers(buf, buf), kong.Exit(func(int) {}))
	_, _ = p.Parse([]string{"--help"})
	// Help from signature should still be present since field tag didn't override it.
	assert.Contains(t, buf.String(), "Signature help")
}

// Non-struct type implementing Signature as a flag.
type signatureFlag string

func (signatureFlag) Signature() string {
	return `help:"Flag from signature" default:"sigdefault"`
}

func TestSignatureNonStructFlag(t *testing.T) {
	var cli struct {
		Flag signatureFlag
	}
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	assert.NoError(t, err)
	assert.Equal(t, signatureFlag("sigdefault"), cli.Flag)
}

func TestSignatureNonStructFlagFieldOverrides(t *testing.T) {
	var cli struct {
		Flag signatureFlag `default:"fieldval"`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse(nil)
	assert.NoError(t, err)
	// Field tag should override the signature default.
	assert.Equal(t, signatureFlag("fieldval"), cli.Flag)
}

// Empty signature should be ignored.
type emptySignatureCmd struct{}

func (emptySignatureCmd) Signature() string { return "" }

func TestSignatureEmptyIgnored(t *testing.T) {
	var cli struct {
		Cmd emptySignatureCmd `cmd:""`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"cmd"})
	assert.NoError(t, err)
}

// Whitespace-only signature should also be ignored.
type whitespaceSignatureCmd struct{}

func (whitespaceSignatureCmd) Signature() string { return "   " }

func TestSignatureWhitespaceIgnored(t *testing.T) {
	var cli struct {
		Cmd whitespaceSignatureCmd `cmd:""`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"cmd"})
	assert.NoError(t, err)
}
