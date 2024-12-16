package kong

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// A Resolver resolves a Flag value from an external source.
type Resolver interface {
	// Validate configuration against Application.
	//
	// This can be used to validate that all provided configuration is valid within  this application.
	Validate(app *Application) error

	// Resolve the value for a Flag.
	Resolve(context *Context, parent *Path, flag *Flag) (interface{}, error)
}

// ResolverFunc is a convenience type for non-validating Resolvers.
type ResolverFunc func(context *Context, parent *Path, flag *Flag) (interface{}, error)

var _ Resolver = ResolverFunc(nil)

func (r ResolverFunc) Resolve(context *Context, parent *Path, flag *Flag) (interface{}, error) { //nolint: revive
	return r(context, parent, flag)
}
func (r ResolverFunc) Validate(app *Application) error { return nil } //nolint: revive

// JSON returns a Resolver that retrieves values from a JSON source.
//
// Flag names are used as JSON keys indirectly, by tring snake_case and camelCase variants.
func JSON(r io.Reader) (Resolver, error) {
	values := map[string]interface{}{}
	err := json.NewDecoder(r).Decode(&values)
	if err != nil {
		return nil, err
	}
	var f ResolverFunc = func(context *Context, parent *Path, flag *Flag) (interface{}, error) {
		name := strings.ReplaceAll(flag.Name, "-", "_")
		snakeCaseName := snakeCase(flag.Name)
		raw, ok := values[name]
		if ok {
			return raw, nil
		} else if raw, ok = values[snakeCaseName]; ok {
			return raw, nil
		}
		raw = values
		for _, part := range strings.Split(name, ".") {
			if values, ok := raw.(map[string]interface{}); ok {
				raw, ok = values[part]
				if !ok {
					return nil, nil
				}
			} else {
				return nil, nil
			}
		}
		return raw, nil
	}

	return f, nil
}

func snakeCase(name string) string {
	name = strings.Join(strings.Split(strings.Title(name), "-"), "")
	return strings.ToLower(name[:1]) + name[1:]
}

func EnvResolver() Resolver {
	once := Once()
	return ResolverFunc(func(context *Context, parent *Path, flag *Flag) (interface{}, error) {
		if err := once(func() error { return visit(context.Path) }); err != nil {
			return nil, err
		}
		for _, env := range flag.Tag.Envs {
			envar, ok := os.LookupEnv(env)
			// Parse the first non-empty ENV in the list
			if ok {
				return envar, nil
			}
		}
		return nil, nil
	})
}

func visit(paths []*Path) error {
	for _, path := range paths {
		if path.Command == nil {
			continue
		}
		for _, positional := range path.Command.Positional {
			if positional.Tag == nil {
				continue
			}
			visitValue(positional)
		}
		if path.Command.Argument != nil {
			visitValue(path.Command.Argument)
		}
	}
	return nil
}

func visitValue(value *Value) error {
	for _, env := range value.Tag.Envs {
		envar, ok := os.LookupEnv(env)
		if !ok {
			continue
		}
		token := Token{Type: FlagValueToken, Value: envar}
		if err := value.Parse(ScanFromTokens(token), value.Target); err != nil {
			return fmt.Errorf("%s (from envar %s=%q)", err, env, envar)
		}
	}
	return nil
}

func Once() func(func() error) error {
	done := false
	return func(fn func() error) error {
		if !done {
			done = true
			return fn()
		}
		return nil
	}
}
