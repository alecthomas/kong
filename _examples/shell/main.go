package main

import (
	"encoding/json"
	"fmt"

	"github.com/alecthomas/kong"
)

var CLI struct {
	Rm struct {
		Force     bool `kong:"help='Force removal.'"`
		Recursive bool `kong:"help='Recursively remove files.'"`

		Paths []string `kong:"help='Paths to remove.',type='path'"`
	} `kong:"help='Remove files.'"`

	Ls struct {
		Paths []string `kong:"help='Paths to list.',type='path'"`
	} `kong:"help='List paths.'"`
}

func main() {
	cmd := kong.Parse(&CLI)
	s, _ := json.Marshal(&CLI)
	fmt.Println(cmd)
	fmt.Println(string(s))
}
