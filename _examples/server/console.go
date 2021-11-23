// nolint: govet
package main

import (
	"fmt"

	"github.com/alecthomas/kong"
)

// Ensure the grammar compiles.
var _ = kong.Must(&grammar{})

// Server interface.
type grammar struct {
	Help     helpCmd `cmd:"" help:"Show help."`
	Question helpCmd `cmd:"" hidden:"" name:"?" help:"Show help."`

	Status statusCmd `cmd:"" help:"Show server status."`
}

type statusCmd struct {
	Verbose bool `short:"v" help:"Show verbose status information."`
}

func (s *statusCmd) Run(ctx *kong.Context) error {
	ctx.Printf("OK")
	return nil
}

type helpCmd struct {
	Command []string `arg:"" optional:"" help:"Show help on command."`
}

// Run shows help.
func (h *helpCmd) Run(realCtx *kong.Context) error {
	ctx, err := kong.Trace(realCtx.Kong, h.Command)
	if err != nil {
		return err
	}
	if ctx.Error != nil {
		return ctx.Error
	}
	err = ctx.PrintUsage(false)
	if err != nil {
		return err
	}
	fmt.Fprintln(realCtx.Stdout)
	return nil
}
