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

		Paths []string `arg:"" help:"Paths to remove." type:"path"`
	} `cmd:"" help:"Remove files."`

	Ls struct {
		Paths []string `arg:"" optional:"" help:"Paths to list." type:"path"`
	} `cmd:"" help:"List paths."`
}

func main() {
	cmd := kong.Parse(&cli, kong.Description("A shell-like example app."), kong.HelpOptions(kong.CompactHelp()))
	switch cmd {
	case "rm <paths>":
		fmt.Println(cli.Rm.Paths, cli.Rm.Force, cli.Rm.Recursive)

	case "ls":
	}
}
