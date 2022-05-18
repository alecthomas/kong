package main

import (
	"github.com/alecthomas/kong"
)

var cli struct {
	DryRun    string `help:"optional dry run flag"`
	Mandatory bool   `required:"" help:"mandatory global flag"`
	Resource  struct {
		Create struct {
			Name string `required:"" help:"name of the resource"`
		} `cmd:"" help:"create a resource"`
		Delete struct {
			Scope   string   `required:"" help:"mandatory subcommand flag"`
			Labels  []string `help:"labels to match when deleting"`
			Version string   `help:"version to match when deleting"`
		} `cmd:"" help:"delete resource(s)"`
	} `cmd:"" help:"operate on resources"`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("help"),
		kong.Description("An app demonstrating SubcommandsWithOptionalFlags"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			SubcommandsWithOptionalFlags: true,
		}))
	ctx.Run()
}
