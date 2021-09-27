package kong

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
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

// Unzip unzips data and writes it to the Writer
func Unzip(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create new gzip reader")
	}
	defer gr.Close()

	data, err = ioutil.ReadAll(gr)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read gzip data")
	}

	return data, nil
}
