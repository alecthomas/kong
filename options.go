package kong

import (
	"io"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"
)

// An Option applies optional changes to the Kong application.
type Option func(k *Kong) error

// Exit overrides the function used to terminate. This is useful for testing or interactive use.
func Exit(exit func(int)) Option {
	return func(k *Kong) error {
		k.Exit = exit
		return nil
	}
}

// NoDefaultHelp disables the default help flags.
func NoDefaultHelp() Option {
	return func(k *Kong) error {
		k.noDefaultHelp = true
		return nil
	}
}

// Name overrides the application name.
func Name(name string) Option {
	return func(k *Kong) error {
		k.postBuildOptions = append(k.postBuildOptions, func(k *Kong) error {
			k.Model.Name = name
			return nil
		})
		return nil
	}
}

// Description sets the application description.
func Description(description string) Option {
	return func(k *Kong) error {
		k.postBuildOptions = append(k.postBuildOptions, func(k *Kong) error {
			k.Model.Help = description
			return nil
		})
		return nil
	}
}

// TypeMapper registers a mapper to a type.
func TypeMapper(typ reflect.Type, mapper Mapper) Option {
	return func(k *Kong) error {
		k.registry.RegisterType(typ, mapper)
		return nil
	}
}

// KindMapper registers a mapper to a kind.
func KindMapper(kind reflect.Kind, mapper Mapper) Option {
	return func(k *Kong) error {
		k.registry.RegisterKind(kind, mapper)
		return nil
	}
}

// ValueMapper registers a mapper to a field value.
func ValueMapper(ptr interface{}, mapper Mapper) Option {
	return func(k *Kong) error {
		k.registry.RegisterValue(ptr, mapper)
		return nil
	}
}

// NamedMapper registers a mapper to a name.
func NamedMapper(name string, mapper Mapper) Option {
	return func(k *Kong) error {
		k.registry.RegisterName(name, mapper)
		return nil
	}
}

// Writers overrides the default writers. Useful for testing or interactive use.
func Writers(stdout, stderr io.Writer) Option {
	return func(k *Kong) error {
		k.Stdout = stdout
		k.Stderr = stderr
		return nil
	}
}

// HookFunc is a callback tied to a field of the grammar, called before a value is applied.
type HookFunc func(ctx *Context, path *Path) error

// Hook to apply before a command, flag or positional argument is encountered.
//
// "ptr" is a pointer to a field of the grammar.
//
// Note that the hook will be called once for each time the corresponding node is encountered. This means that if a flag
// is passed twice, its hook will be called twice.
func Hook(ptr interface{}, hook HookFunc) Option {
	key := reflect.ValueOf(ptr)
	if key.Kind() != reflect.Ptr {
		panic("expected a pointer")
	}
	return func(k *Kong) error {
		k.before[key] = hook
		return nil
	}
}

// HelpFunction is the type of a function used to display help.
type HelpFunction func(*Context) error

// Help function to use.
//
// Defaults to PrintHelp.
func Help(help HelpFunction) Option {
	return func(k *Kong) error {
		k.help = help
		return nil
	}
}

// HelpOptions specifies options for the default help printer, if used.
func HelpOptions(options ...HelpOption) Option {
	return func(k *Kong) error {
		k.helpOptions = options
		return nil
	}
}

// ClearResolvers clears all existing resolvers.
func ClearResolvers() Option {
	return func(k *Kong) error {
		k.resolvers = nil
		return nil
	}
}

// Resolver registers flag resolvers.
func Resolver(resolvers ...ResolverFunc) Option {
	return func(k *Kong) error {
		k.resolvers = append(k.resolvers, resolvers...)
		return nil
	}
}

// ConfigurationFunc is a function that builds a resolver from a file.
type ConfigurationFunc func(r io.Reader) (ResolverFunc, error)

// Configuration provides Kong with support for loading defaults from a set of configuration files.
//
// Paths will be opened in order, and "loader" will be used to provide a ResolverFunc which is registered with Kong.
//
// Note: The JSON function is a ConfigurationFunc.
//
// ~ expansion will occur on the provided paths.
func Configuration(loader ConfigurationFunc, paths ...string) Option {
	return func(k *Kong) error {
		for _, path := range paths {
			path = expandPath(path)
			r, err := os.Open(path) // nolint: gas
			if err != nil {
				continue
			}
			resolver, err := loader(r)
			if err == nil {
				k.resolvers = append(k.resolvers, resolver)
			}
			_ = r.Close()
		}
		return nil
	}
}

func expandPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if strings.HasPrefix(path, "~/") {
		user, err := user.Current()
		if err != nil {
			return path
		}
		return filepath.Join(user.HomeDir, path[2:])
	}
	abspath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abspath
}
