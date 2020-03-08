package kong

import (
	"github.com/posener/complete"
)

// Completers contains custom Completers used to generate completion options.
//
// They can be used with an annotation like `completer='myCustomCompleter'` where "myCustomCompleter" is a
// key in Completers.
type Completers map[string]Completer

// Apply completers to Kong as a configuration option.
func (c Completers) Apply(k *Kong) error {
	k.completers = c
	return nil
}

func defaultCompleter(ctx *Context) (bool, error) {
	cmd, err := nodeCompleteCommand(ctx.Model.Node, ctx.completers)
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
