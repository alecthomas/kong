package kong

import (
	"github.com/posener/complete/cmd/install"
)

// CompletionOptions options for shell completion.
type CompletionOptions struct {
	// the function to run for shell completions. If nil, no completions will be run. default: nil
	Completer Completer
}

// InstallCompletion will install completion to your shell
type InstallCompletion struct {
	Installer func(*Context) error
}

//Run runs install
func (c *InstallCompletion) Run(ctx *Context) error {
	if c.Installer != nil {
		return c.Installer(ctx)
	}
	// use installer from github.com/posener/complete by default
	return install.Install(ctx.Model.Name)
}

// UninstallCompletion will uninstall completion from your shell (reverses InstallCompletion)
type UninstallCompletion struct {
	Uninstaller func(*Context) error
}

//Run runs uninstall
func (c *UninstallCompletion) Run(ctx *Context) error {
	if c.Uninstaller != nil {
		return c.Uninstaller(ctx)
	}
	// use uninstaller from github.com/posener/complete by default
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

func runCompletion(ctx *Context, completer Completer, exit func(int)) error {
	if completer == nil {
		return nil
	}
	ran, err := completer(ctx)
	if err != nil {
		return err
	}
	if ran {
		exit(0)
	}
	return nil
}
