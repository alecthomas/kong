package main

import (
	"fmt"

	"github.com/alecthomas/kong"
)

var cli struct {
	Flag flagWithHelp    `help:"${flag_help}"`
	Echo commandWithHelp `cmd:"" help:"Regular command help"`
}

type flagWithHelp bool

// See https://github.com/alecthomas/kong?tab=readme-ov-file#variable-interpolation
var vars = kong.Vars{
	"flag_help": "Extended flag help that might be too long for directly " +
		"including in the struct tag field",
}

type commandWithHelp struct {
	Msg argumentWithHelp `arg:"" help:"Regular argument help"`
}

func (c *commandWithHelp) Help() string {
	return "🚀 additional command help"
}

type argumentWithHelp struct {
	Msg string `arg:""`
}

func (f *argumentWithHelp) Help() string {
	return "📣 additional argument help"
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("help"),
		kong.Description("An app demonstrating HelpProviders"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: false,
		}),
		vars)
	switch ctx.Command() {
	case "echo <msg>":
		fmt.Println(cli.Echo.Msg)
	}
}
