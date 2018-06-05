package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

// nolint: govet
var CLI struct {
	Debug bool `help:"Debug mode."`

	Rm struct {
		User      string `help:"Run as user." short:"u"`
		Force     bool   `help:"Force removal." short:"f"`
		Recursive bool   `help:"Recursively remove files." short:"r"`

		Paths []string `arg help:"Paths to remove." type:"path"`
	} `cmd help:"Remove files."`

	Ls struct {
		Paths []string `arg optional help:"Paths to list." type:"path"`
	} `cmd help:"List paths."`
}

func main() {
	app := kong.Must(&CLI, kong.Description("A shell-like example app."))
	cmd, err := app.Parse(os.Args[1:])
	app.FatalIfErrorf(err)
	s, _ := json.Marshal(&CLI)
	fmt.Println(cmd)
	fmt.Println(string(s))
}
