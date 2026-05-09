package main

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"
)

const logo = `
#   #   ###    #   #    ###
#  #   #   #   ##  #   #
###    #   #   # # #   #  ##
#  #   #   #   #  ##   #   #
#   #   ###    #   #    ###`

var cli struct {
	Quiet bool `help:"Suppress non-error output."`

	Say struct {
		Text string `arg:"" help:"Text to echo."`
	} `cmd:"" help:"Print text to stdout."`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("kong-logo"),
		kong.Description(strings.TrimSpace(logo)+"\n\n"+
			"Simple example which shows how to disable formatting of\n"+
			"the description text, to display ASCII art as-is."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			NoAppDescFormat: true,
		}),
	)
	switch ctx.Command() {
	case "say <text>":
		if !cli.Quiet {
			fmt.Println(cli.Say.Text)
		}
	}
}
