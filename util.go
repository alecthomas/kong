package kong

import (
	"fmt"
	"os"
	"reflect"
)

// ConfigFlag uses the configured (via kong.Configuration(loader)) configuration loader to load configuration
// from a file specified by a flag.
//
// Use this as a flag value to support loading of custom configuration via a flag.
type ConfigFlag string

// BeforeResolve adds a resolver.
func (c ConfigFlag) BeforeResolve(kong *Kong, ctx *Context, trace *Path) (err error) {
	var (
		path     string
		resolver Resolver
	)

	if kong.loader == nil {
		err = Errors().KongMustBeConfiguredWithConfiguration()
		return
	}
	path = string(ctx.FlagValue(trace.Flag).(ConfigFlag))
	if resolver, err = kong.LoadConfig(path); err != nil {
		return
	}
	ctx.AddResolver(resolver)

	return
}

// VersionFlag is a flag type that can be used to display a version number, stored in the "version" variable.
type VersionFlag bool

// BeforeApply writes the version variable and terminates with a 0 exit status.
func (v VersionFlag) BeforeApply(app *Kong, vars Vars) error {
	_, _ = fmt.Fprintln(app.Stdout, vars[keyVersion])
	app.Exit(0)
	return nil
}

// ChangeDirFlag changes the current working directory to a path specified by a flag
// early in the parsing process, changing how other flags resolve relative paths.
//
// Use this flag to provide a "git -C" like functionality.
//
// It is not compatible with custom named decoders, e.g., existingdir.
type ChangeDirFlag string

// Decode is used to create a side effect of changing the current working directory.
func (c ChangeDirFlag) Decode(ctx *DecodeContext) (err error) {
	var path string

	if err = ctx.Scan.PopValueInto(keyString, &path); err != nil {
		return
	}
	path = ExpandPath(path)
	ctx.Value.Target.Set(reflect.ValueOf(ChangeDirFlag(path)))
	err = os.Chdir(path)

	return
}
