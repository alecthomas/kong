package kong

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// ResolverFunc resolves a Flag value from an external source.
type ResolverFunc func(flag *Flag) (string, error)

// JSONResolver returns a Resolver that retrieves values from a JSON source.
func JSONResolver(r io.Reader) (ResolverFunc, error) {
	values := map[string]interface{}{}
	err := json.NewDecoder(r).Decode(&values)
	if err != nil {
		return nil, err
	}
	f := func(flag *Flag) (string, error) {
		raw, ok := values[flag.Name]
		if !ok {
			return "", nil
		}
		value, err := jsonDecodeValue(flag.Tag.Sep, raw)
		if err != nil {
			return "", err
		}
		return value, nil
	}

	return f, nil
}

func jsonDecodeValue(sep string, value interface{}) (string, error) {
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
		return strings.Join(out, sep), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	}
	return "", fmt.Errorf("unsupported JSON value %v (of type %T)", value, value)
}

// EnvResolver automatically determines environment variables based on the name of each flag, transformed to uppercase
// and underscored, e.g. `my-flag` -> `MY_FLAG` The environment variable key can be overridden with the `env` tag.
func EnvResolver(prefix string) (ResolverFunc, error) {
	f := func(flag *Flag) (string, error) {
		v, _ := os.LookupEnv(envString(prefix, flag))
		return v, nil
	}
	return f, nil
}

func envString(prefix string, flag *Flag) string {
	if env, ok := flag.Tag.Get("env"); ok {
		return env
	}

	env := strings.ToUpper(flag.Name)
	env = strings.Replace(env, "-", "_", -1)
	env = prefix + env

	return env
}
