package kong

import (
	"fmt"
)

// ConfigFlag uses the configured (via kong.Configuration(loader)) configuration loader to load configuration
// from a file specified by a flag.
//
// Use this as a flag value to support loading of custom configuration via a flag.
type ConfigFlag string

// BeforeResolve adds a resolver.
func (c ConfigFlag) BeforeResolve(kong *Kong, ctx *Context, trace *Path) error {
	if kong.loader == nil {
		return fmt.Errorf("kong must be configured with kong.Configuration(...)")
	}
	path := string(ctx.FlagValue(trace.Flag).(ConfigFlag))
	resolver, err := kong.LoadConfig(path)
	if err != nil {
		return err
	}
	ctx.AddResolver(resolver)
	return nil
}

// VersionFlag is a flag type that can be used to display a version number, stored in the "version" variable.
type VersionFlag bool

// BeforeApply writes the version variable and terminates with a 0 exit status.
func (v VersionFlag) BeforeApply(app *Kong, vars Vars) error {
	fmt.Fprintln(app.Stdout, vars["version"])
	app.Exit(0)
	return nil
}

// DebugConfigFlag displays the parsed configuration files defined via kong.Configuration(loader).
//
// Use this flag to determine which files got parsed in which order.
type DebugConfigFlag bool

// BeforeApply writes the parsed configuration files to stderr.
func (d DebugConfigFlag) BeforeApply(app *Kong, ctx *Context, path *Path) error {
	show := bool(ctx.FlagValue(path.Flag).(DebugConfigFlag))
	if show {
		fmt.Fprintln(app.Stderr, "Parsed configuration files:")
		for _, p := range app.configs {
			fmt.Fprintln(app.Stderr, SpaceIndenter("")+p)
		}
	}
	return nil
}
