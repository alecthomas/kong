package kong

import (
	"github.com/posener/complete"
	"github.com/posener/complete/cmd/install"
)

// CompletionOptions options for shell completion.
type CompletionOptions struct {
	// Predictors contains custom Predictors used to generate completion options.
	// They can be used with an annotation like `predictor='myCustomPredictor'` where "myCustomPredictor" is a
	// key in Predictors.
	Predictors map[string]Predictor
}

// InstallCompletion will install completion to your shell
type InstallCompletion struct{}

func (c *InstallCompletion) IsBool() bool { return true }

func (c *InstallCompletion) Decode(ctx *DecodeContext) error { return nil }

// Run runs install.
func (c *InstallCompletion) Run(ctx *Context) error {
	return install.Install(ctx.Model.Name)
}

// UninstallCompletion will uninstall completion from your shell (reverses InstallCompletion)
type UninstallCompletion struct{}

// Run runs uninstall
func (c *UninstallCompletion) Run(ctx *Context) error {
	return install.Uninstall(ctx.Model.Name)
}

// Completer is a function to run shell completions. Returns true if this was a completion run. Kong will exit 0
// immediately when it returns true.
type Completer func(*Context) (bool, error)

// Apply options to Kong as a configuration option.
func (c CompletionOptions) Apply(k *Kong) error {
	k.completionOptions = c
	return nil
}

func defaultCompleter(ctx *Context) (bool, error) {
	cmd, err := nodeCompleteCommand(ctx.Model.Node, ctx.completionOptions.Predictors)
	if err != nil {
		return false, err
	}
	cmp := complete.New(ctx.Kong.Model.Name, *cmd)
	cmp.Out = ctx.Kong.Stdout
	return cmp.Complete(), nil
}

// nodeCompleteCommand recursively builds a *complete.Command for a node and all of its dear children.
func nodeCompleteCommand(node *Node, predictors map[string]Predictor) (*complete.Command, error) {
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
		childCmd, err := nodeCompleteCommand(child, predictors)
		if err != nil {
			return nil, err
		}
		if childCmd != nil {
			cmd.Sub[child.Name] = *childCmd
		}
	}

	var err error
	cmd.GlobalFlags, err = nodeGlobalFlags(node, predictors)
	if err != nil {
		return nil, err
	}

	pps, err := positionalPredictors(node.Positional, predictors)
	if err != nil {
		return nil, err
	}
	cmd.Args = newCompletePredictor(&positionalPredictor{
		Predictors: pps,
		Flags:      node.Flags,
	})

	return &cmd, nil
}

func nodeGlobalFlags(node *Node, predictors map[string]Predictor) (map[string]complete.Predictor, error) {
	if node == nil || node.Flags == nil {
		return map[string]complete.Predictor{}, nil
	}
	globalFlags := make(map[string]complete.Predictor, len(node.Flags)*2)
	for _, flag := range node.Flags {
		if flag == nil {
			continue
		}
		predictor, err := flagPredictor(flag, predictors)
		if err != nil {
			return nil, err
		}
		cp := newCompletePredictor(predictor)
		globalFlags["--"+flag.Name] = cp
		if flag.Short == 0 {
			continue
		}
		globalFlags["-"+string(flag.Short)] = cp
	}
	return globalFlags, nil
}
