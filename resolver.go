package kong

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type ResolverFunc func(flag *Flag) (string, error)

// JSONResolver returns a Resolver that retrieves values from a JSON source.
func JSONResolver(r io.Reader) (ResolverFunc, error) {
	values := map[string]interface{}{}
	err := json.NewDecoder(r).Decode(&values)
	if err != nil {
		return nil, err
	}
	mapping := map[string]string{}
	for key, value := range values {
		sub, err := jsonDecodeValue(value)
		if err != nil {
			return nil, err
		}
		mapping[key] = sub
	}

	f := func(flag *Flag) (string, error) {
		return mapping[flag.Name], nil
	}

	return f, nil
}

func jsonDecodeValue(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case float64:
		return fmt.Sprintf("%v", v), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	}
	return "", fmt.Errorf("unsupported JSON value %v (of type %T)", value, value)
}

// Automatically determines environment variables based on the name of each flag,
// transformed to uppercase and underscored, e.g. `my-flag` -> `MY_FLAG`
// The environment variable key can be overridden with the `env` tag.
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
