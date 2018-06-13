package kong

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// ResolverFunc resolves a Flag value from an external source.
type ResolverFunc func(context *Context, parent *Path, flag *Flag) (string, error)

// JSON returns a Resolver that retrieves values from a JSON source.
//
// Hyphens in flag names are replaced with underscores.
func JSON(r io.Reader) (ResolverFunc, error) {
	values := map[string]interface{}{}
	err := json.NewDecoder(r).Decode(&values)
	if err != nil {
		return nil, err
	}
	f := func(context *Context, parent *Path, flag *Flag) (string, error) {
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

// Envars resolves flag values using the `env:"<name>"` tag. It ignores flags without this tag.
//
// This resolver is installed by default.
func Envars() ResolverFunc {
	return func(context *Context, parent *Path, flag *Flag) (string, error) {
		if flag.Tag.Env == "" {
			return "", nil
		}
		v, _ := os.LookupEnv(flag.Tag.Env)
		return v, nil
	}
}
