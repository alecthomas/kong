package kong

import (
	"reflect"
)

//CompletionOptions options for shell completion.
type CompletionOptions struct {
	// the function to run for shell completions. If nil, no completions will be run. default: nil
	Completer Completer

	// installer/uninstaller for shell completions.  If nil, there will be no install or uninstall flags
	CompletionInstaller CompletionInstaller

	// flag name for installing completions to a shell.  default: --install-shell-completion
	InstallFlagName string

	// help text for --install-shell-completion. default: "Install shell completion."
	InstallFlagHelp string

	// flag name for uninstalling completions from a shell.  default: --uninstall-shell-completion
	UninstallFlagName string

	// help text for --uninstall-shell-completion. default: "Uninstall shell completion."
	UninstallFlagHelp string

	// whether to hide --install-shell-completion and --uninstall-shell-completion from help.  default: false
	HideFlags bool
}

func (c CompletionOptions) withDefaults() CompletionOptions {
	if c.InstallFlagName == "" {
		c.InstallFlagName = "install-shell-completion"
	}
	if c.UninstallFlagName == "" {
		c.UninstallFlagName = "uninstall-shell-completion"
	}
	if c.InstallFlagHelp == "" {
		c.InstallFlagHelp = "Install shell completion."
	}
	if c.UninstallFlagHelp == "" {
		c.UninstallFlagHelp = "Uninstall shell completion."
	}
	return c
}

// Completer is a function to run shell completions. Returns true if this was a completion run. Kong will exit 0
// immediately when it returns true.
type Completer func(*Context) (bool, error)

//CompletionInstaller contains functions to install completions from to a shell.
type CompletionInstaller interface {
	Install(ctx *Context) error
	Uninstall(ctx *Context) error
}

// Apply options to Kong as a configuration option.
func (c CompletionOptions) Apply(k *Kong) error {
	k.completionOptions = c
	return nil
}

type installCompletionsValue bool

func (b installCompletionsValue) BeforeApply(ctx *Context) error {
	inst := ctx.Kong.completionOptions.CompletionInstaller
	if inst != nil {
		err := inst.Install(ctx)
		if err != nil {
			return err
		}
	}
	ctx.Kong.Exit(0)
	return nil
}

type uninstallCompletionsValue bool

func (b uninstallCompletionsValue) BeforeApply(ctx *Context) error {
	inst := ctx.Kong.completionOptions.CompletionInstaller
	if inst != nil {
		err := inst.Uninstall(ctx)
		if err != nil {
			return err
		}
	}
	ctx.Kong.Exit(0)
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

func completionFlags(k *Kong) []*Flag {
	options := k.completionOptions
	if options.CompletionInstaller == nil {
		return []*Flag{}
	}
	options = options.withDefaults()

	var instTarget installCompletionsValue
	instVal := reflect.ValueOf(&instTarget).Elem()
	instFlag := &Flag{
		Hidden: options.HideFlags,
		Value: &Value{
			Name:         options.InstallFlagName,
			Help:         options.InstallFlagHelp,
			Target:       instVal,
			Tag:          &Tag{},
			Mapper:       k.registry.ForValue(instVal),
			DefaultValue: reflect.ValueOf(false),
		},
	}
	instFlag.Flag = instFlag
	k.instCompletionFlag = instFlag

	var uninstTarget uninstallCompletionsValue
	uninstVal := reflect.ValueOf(&uninstTarget).Elem()
	uninstFlag := &Flag{
		Hidden: options.HideFlags,
		Value: &Value{
			Name:         options.UninstallFlagName,
			Help:         options.UninstallFlagHelp,
			Target:       uninstVal,
			Tag:          &Tag{},
			Mapper:       k.registry.ForValue(uninstVal),
			DefaultValue: reflect.ValueOf(false),
		},
	}
	uninstFlag.Flag = uninstFlag
	k.uninstCompletionFlag = uninstFlag

	return []*Flag{instFlag, uninstFlag}
}
