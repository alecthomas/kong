package kong

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// A Resolver resolves a Flag value from an external source.
type Resolver interface {
	// Validate configuration against Application.
	//
	// This can be used to validate that all provided configuration is valid within  this application.
	Validate(app *Application) error

	// Resolve the value for a Flag.
	Resolve(context *Context, parent *Path, flag *Flag) (string, error)
}

// ResolverFunc is a convenience type for non-validating Resolvers.
type ResolverFunc func(context *Context, parent *Path, flag *Flag) (string, error)

var _ Resolver = ResolverFunc(nil)

func (r ResolverFunc) Resolve(context *Context, parent *Path, flag *Flag) (string, error) { // nolint: golint
	return r(context, parent, flag)
}
func (r ResolverFunc) Validate(app *Application) error { return nil } //  nolint: golint

// JSON returns a Resolver that retrieves values from a JSON source.
//
// Hyphens in flag names are replaced with underscores.
func JSON(r io.Reader) (Resolver, error) {
	values := map[string]interface{}{}
	err := json.NewDecoder(r).Decode(&values)
	if err != nil {
		return nil, err
	}
	var f ResolverFunc = func(context *Context, parent *Path, flag *Flag) (string, error) {
		name := strings.Replace(flag.Name, "-", "_", -1)
		raw, ok := values[name]
		if !ok {
			return "", nil
		}
		sep := flag.Tag.Sep
		value, err := jsonDecodeValue(sep, raw)
		if err != nil {
			return "", err
		}
		return value, nil
	}

	return f, nil
}

func jsonDecodeValue(sep rune, value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case float64:
		return fmt.Sprintf("%v", v), nil
	case []interface{}:
		out := []string{}
		for _, el := range v {
			sel, err := jsonDecodeValue(sep, el)
			if err != nil {
				return "", err
			}
			out = append(out, sel)
		}
		return JoinEscaped(out, sep), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	}
	return "", fmt.Errorf("unsupported JSON value %v (of type %T)", value, value)
}
