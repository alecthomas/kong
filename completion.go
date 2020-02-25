package kong

import (
	"github.com/posener/complete/cmd/install"
)

// CompletionOptions options for shell completion.
type CompletionOptions struct {
	// Completer is the function to run for shell completions. If nil, no completions will be run. default: nil
	Completer Completer

	// CompletionInstaller determines the functions to run when running InstallCompletion and UninstallCompletion.
	// By default, Install and Uninstall functions from github.com/posener/complete/cmd/install are used.
	CompletionInstaller CompletionInstaller
}

// CompletionInstaller contains functions to install completions from to a shell.
type CompletionInstaller interface {
	Install(ctx *Context) error
	Uninstall(ctx *Context) error
}

// NewCompletionInstaller returns a CompletionInstaller with the given install and uninstall funcs
// leaving either nil will cause it to defer to the default installer/uninstaller
func NewCompletionInstaller(install, uninstall func(*Context) error) CompletionInstaller {
	return &completionInstaller{
		install:   install,
		uninstall: uninstall,
	}
}

// completionInstaller is an implementation of CompletionInstaller
type completionInstaller struct {
	install   func(ctx *Context) error
	uninstall func(ctx *Context) error
}

func (c *completionInstaller) Install(ctx *Context) error {
	if c.install != nil {
		return c.install(ctx)
	}
	return defaultCompletionInstall(ctx)
}

func defaultCompletionInstall(ctx *Context) error {
	return install.Install(ctx.Model.Name)
}

func (c *completionInstaller) Uninstall(ctx *Context) error {
	if c.uninstall != nil {
		return c.uninstall(ctx)
	}
	return defaultCompletionUninstall(ctx)
}

func defaultCompletionUninstall(ctx *Context) error {
	return install.Uninstall(ctx.Model.Name)
}

// InstallCompletion will install completion to your shell
type InstallCompletion struct{}

//Run runs install
func (c *InstallCompletion) Run(ctx *Context) error {
	inst := ctx.Kong.completionOptions.CompletionInstaller
	if inst == nil {
		inst = NewCompletionInstaller(nil, nil)
	}
	return inst.Install(ctx)
}

// UninstallCompletion will uninstall completion from your shell (reverses InstallCompletion)
type UninstallCompletion struct{}

//Run runs uninstall
func (c *UninstallCompletion) Run(ctx *Context) error {
	inst := ctx.Kong.completionOptions.CompletionInstaller
	if inst == nil {
		inst = NewCompletionInstaller(nil, nil)
	}
	return inst.Uninstall(ctx)
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
