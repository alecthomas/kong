package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

var CLI struct {
	Help bool `kong:"help='Display help.'"`
	Rm   struct {
		Force     bool `kong:"help='Force removal.'"`
		Recursive bool `kong:"help='Recursively remove files.'"`

		Paths []string `kong:"arg,help='Paths to remove.',type='path'"`
	} `kong:"cmd,help='Remove files.'"`

	Ls struct {
		Paths []string `kong:"help='Paths to list.',type='path'"`
	} `kong:"cmd,help='List paths.'"`
}

func main() {
	app := kong.Must(&CLI).Hook(&CLI.Help, kong.Help(nil, nil))
	cmd, err := app.Parse(os.Args[1:])
	app.FatalIfErrorf(err)
	s, _ := json.Marshal(&CLI)
	fmt.Println(cmd)
	fmt.Println(string(s))
}
