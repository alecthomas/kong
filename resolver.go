package kong

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// A Resolver resolves a Flag value from an external source.
type Resolver interface {
	// Validate configuration against Application.
	//
	// This can be used to validate that all provided configuration is valid within  this application.
	Validate(app *Application) error

	// Resolve the value for a Flag.
	Resolve(context *Context, parent *Path, flag *Flag) (any, error)
}

// NamedResolver can be implemented by a Resolver to identify itself for value provenance.
type NamedResolver interface {
	Name() string
}

// ResolverFunc is a convenience type for non-validating Resolvers.
type ResolverFunc func(context *Context, parent *Path, flag *Flag) (any, error)

var _ Resolver = ResolverFunc(nil)

func (r ResolverFunc) Resolve(context *Context, parent *Path, flag *Flag) (any, error) { //nolint: revive
	return r(context, parent, flag)
}
func (r ResolverFunc) Validate(app *Application) error { return nil } //nolint: revive

// NamedResolverFunc creates a Resolver with a custom name from a function.
func NamedResolverFunc(name string, fn func(context *Context, parent *Path, flag *Flag) (any, error)) Resolver {
	return namedResolverFunc{name: name, fn: fn}
}

type namedResolverFunc struct {
	name string
	fn   func(context *Context, parent *Path, flag *Flag) (any, error)
}

func (r namedResolverFunc) Validate(app *Application) error { return nil }
func (r namedResolverFunc) Resolve(context *Context, parent *Path, flag *Flag) (any, error) {
	return r.fn(context, parent, flag)
}
func (r namedResolverFunc) Name() string   { return r.name }
func (r namedResolverFunc) String() string { return r.name }

// WithResolverName wraps an existing Resolver with a custom name for provenance tracking.
func WithResolverName(name string, resolver Resolver) Resolver {
	return namedResolverWrapper{name: name, Resolver: resolver}
}

type namedResolverWrapper struct {
	Resolver
	name string
}

func (w namedResolverWrapper) Name() string   { return w.name }
func (w namedResolverWrapper) String() string { return w.name }

func resolverName(r Resolver) string {
	if r == nil {
		return ""
	}
	if nr, ok := r.(NamedResolver); ok {
		return nr.Name()
	}
	if nr, ok := r.(interface{ ResolverName() string }); ok {
		return nr.ResolverName()
	}
	if s, ok := r.(fmt.Stringer); ok {
		return s.String()
	}
	t := reflect.TypeOf(r)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Name() != "" {
		return t.Name()
	}
	return "resolver"
}

// JSON returns a Resolver that retrieves values from a JSON source.
//
// Flag names are used as JSON keys indirectly, by trying snake_case and camelCase variants.
func JSON(r io.Reader) (Resolver, error) {
	values := map[string]any{}
	err := json.NewDecoder(r).Decode(&values)
	if err != nil {
		return nil, err
	}
	return &jsonResolver{values: values}, nil
}

type jsonResolver struct {
	values map[string]any
}

func (j *jsonResolver) Validate(app *Application) error { return nil }

func (j *jsonResolver) Name() string { return "json" }

func (j *jsonResolver) String() string { return "json" }

func (j *jsonResolver) Resolve(context *Context, parent *Path, flag *Flag) (any, error) {
	name := strings.ReplaceAll(flag.Name, "-", "_")
	snakeCaseName := snakeCase(flag.Name)
	raw, ok := j.values[name]
	if ok {
		return raw, nil
	} else if raw, ok = j.values[snakeCaseName]; ok {
		return raw, nil
	}
	raw = j.values
	for _, part := range strings.Split(name, ".") {
		if values, ok := raw.(map[string]any); ok {
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

func snakeCase(name string) string {
	name = strings.Join(strings.Split(strings.Title(name), "-"), "") //nolint:staticcheck // Unicode punctuation not an issue
	return strings.ToLower(name[:1]) + name[1:]
}
