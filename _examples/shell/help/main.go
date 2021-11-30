package main

import (
	"fmt"

	"github.com/alecthomas/kong"
)

var cli struct {
	Flag flagWithHelp    `help:"Regular flag help"`
	Echo commandWithHelp `cmd:"" help:"Regular command help"`
}

type flagWithHelp bool

func (f *flagWithHelp) Help() string {
	return "🏁 additional flag help"
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
		}))
	switch ctx.Command() {
	case "echo <msg>":
		fmt.Println(cli.Echo.Msg)
	}
}
