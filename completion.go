package kong

import (
	"github.com/posener/complete"
	"github.com/posener/complete/cmd/install"
)

// CompletionOptions options for shell completion.
type CompletionOptions struct {
	// Completers contains custom Completers used to generate completion options.
	// They can be used with an annotation like `completer='myCustomCompleter'` where "myCustomCompleter" is a
	// key in Completers.
	Completers map[string]Completer
}

// Install shell completion for the given command.
func Install(app *Application) error { return install.Install(app.Name) }

// Uninstall complete command given: cmd: is the command name
func Uninstall(app *Application) error { return install.Uninstall(app.Name) }

// InstallCompletionFlag will install completion to your shell
type InstallCompletionFlag bool

// BeforeApply uninstalls completion into the users shell.
func (c *InstallCompletionFlag) BeforeApply(ctx *Context) error {
	err := Install(ctx.Model)
	if err != nil {
		return err
	}
	ctx.Exit(0)
	return nil
}

// UninstallCompletionFlag will uninstall completion from your shell (reverses InstallCompletionFlag)
type UninstallCompletionFlag bool

// BeforeApply uninstalls completion from the users shell.
func (c *UninstallCompletionFlag) BeforeApply(ctx *Context) error {
	err := Uninstall(ctx.Model)
	if err != nil {
		return err
	}
	ctx.Exit(0)
	return nil
}

// Apply options to Kong as a configuration option.
func (c CompletionOptions) Apply(k *Kong) error {
	k.completionOptions = c
	return nil
}

func defaultCompleter(ctx *Context) (bool, error) {
	cmd, err := nodeCompleteCommand(ctx.Model.Node, ctx.completionOptions.Completers)
	if err != nil {
		return false, err
	}
	cmp := complete.New(ctx.Kong.Model.Name, *cmd)
	cmp.Out = ctx.Kong.Stdout
	return cmp.Complete(), nil
}

// nodeCompleteCommand recursively builds a *complete.Command for a node and all of its dear children.
func nodeCompleteCommand(node *Node, completers map[string]Completer) (*complete.Command, error) {
	if node == nil {
		return nil, nil
	}

	cmd := complete.Command{
		Sub:         complete.Commands{},
		GlobalFlags: complete.Flags{},
	}
	for _, child := range node.Children {
		if child == nil {
			continue
		}
		childCmd, err := nodeCompleteCommand(child, completers)
		if err != nil {
			return nil, err
		}
		if childCmd != nil {
			cmd.Sub[child.Name] = *childCmd
		}
	}

	var err error
	cmd.GlobalFlags, err = nodeGlobalFlags(node, completers)
	if err != nil {
		return nil, err
	}

	pps, err := positionalCompleters(node.Positional, completers)
	if err != nil {
		return nil, err
	}
	cmd.Args = &positionalCompleter{
		Completers: pps,
		Flags:      node.Flags,
	}

	return &cmd, nil
}

func nodeGlobalFlags(node *Node, completers map[string]Completer) (map[string]Completer, error) {
	if node == nil || node.Flags == nil {
		return map[string]Completer{}, nil
	}
	globalFlags := make(map[string]Completer, len(node.Flags)*2)
	for _, flag := range node.Flags {
		if flag == nil {
			continue
		}
		completer, err := flagCompleter(flag, completers)
		if err != nil {
			return nil, err
		}
		globalFlags["--"+flag.Name] = completer
		if flag.Short == 0 {
			continue
		}
		globalFlags["-"+string(flag.Short)] = completer
	}
	return globalFlags, nil
}
