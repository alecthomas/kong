package kong

import (
	"os"
)

// EnvarResolver resolves values from environment variables.
//
// It is installed by default. Use ClearResolvers() to disable this.
func EnvarResolver() Resolver {
	return ResolverFunc(func(context *Context, parent *Path, flag *Flag) (interface{}, error) {
		if flag.Tag.Env == "" {
			return nil, nil
		}
		envar := os.Getenv(flag.Tag.Env)
		if envar != "" {
			return envar, nil
		}
		return nil, nil
	})
}
