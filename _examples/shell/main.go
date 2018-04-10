package main

import "github.com/alecthomas/kong"

var CLI struct {
	Rm struct {
		Force     bool `help:"Force removal."`
		Recursive bool `help:"Recursively remove files."`

		Paths []string `help:"Paths to remove." type:"path"`
	} `help:"Remove files."`

	Ls struct {
		Paths []string `help:"Paths to list." type:"path"`
	} `help:"List paths."`
}

func main() {
	kong.Parse(&CLI)
}
