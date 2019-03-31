package main

import (
	"fmt"
	"github.com/alecthomas/kong"
)

var cli struct {
	Debug bool `help:"Debug mode."`

	Rm struct {
		User      string `help:"Run as user." short:"u" default:"default"`
		Force     bool   `help:"Force removal." short:"f"`
		Recursive bool   `help:"Recursively remove files." short:"r"`

		Paths []string `arg:"" help:"Paths to remove." type:"path" name:"path"`
	} `cmd:"" help:"Remove files."`

	Ls struct {
		Paths []string `arg:"" optional:"" help:"Paths to list." type:"path"`
	} `cmd:"" help:"List paths."`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("shell"),
		kong.Description("A shell-like example app."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))
	switch ctx.Command() {
	case "rm <path>":
		fmt.Println(cli.Rm.Paths, cli.Rm.Force, cli.Rm.Recursive)

	case "ls":
	}
}
