package kong

import (
	"io"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"
)

// An Option applies optional changes to the Kong application.
type Option interface {
	Apply(k *Kong) error
}

// OptionFunc is function that adheres to the Option interface.
type OptionFunc func(k *Kong) error

func (o OptionFunc) Apply(k *Kong) error { return o(k) } // nolint: golint

// Vars sets the variables to use for interpolation into help strings and default values.
//
// See README for details.
type Vars map[string]string

// Apply lets Vars act as an Option.
func (v Vars) Apply(k *Kong) error {
	for key, value := range v {
		k.vars[key] = value
	}
	return nil
}

// CloneWith clones the current Vars and merges "vars" onto the clone.
func (v Vars) CloneWith(vars Vars) Vars {
	out := Vars{}
	for key, value := range v {
		out[key] = value
	}
	for key, value := range vars {
		out[key] = value
	}
	return out
}

// Exit overrides the function used to terminate. This is useful for testing or interactive use.
func Exit(exit func(int)) Option {
	return OptionFunc(func(k *Kong) error {
		k.Exit = exit
		return nil
	})
}

// NoDefaultHelp disables the default help flags.
func NoDefaultHelp() Option {
	return OptionFunc(func(k *Kong) error {
		k.noDefaultHelp = true
		return nil
	})
}

// Name overrides the application name.
func Name(name string) Option {
	return OptionFunc(func(k *Kong) error {
		k.postBuildOptions = append(k.postBuildOptions, OptionFunc(func(k *Kong) error {
			k.Model.Name = name
			return nil
		}))
		return nil
	})
}

// Description sets the application description.
func Description(description string) Option {
	return OptionFunc(func(k *Kong) error {
		k.postBuildOptions = append(k.postBuildOptions, OptionFunc(func(k *Kong) error {
			k.Model.Help = description
			return nil
		}))
		return nil
	})
}

// TypeMapper registers a mapper to a type.
func TypeMapper(typ reflect.Type, mapper Mapper) Option {
	return OptionFunc(func(k *Kong) error {
		k.registry.RegisterType(typ, mapper)
		return nil
	})
}

// KindMapper registers a mapper to a kind.
func KindMapper(kind reflect.Kind, mapper Mapper) Option {
	return OptionFunc(func(k *Kong) error {
		k.registry.RegisterKind(kind, mapper)
		return nil
	})
}

// ValueMapper registers a mapper to a field value.
func ValueMapper(ptr interface{}, mapper Mapper) Option {
	return OptionFunc(func(k *Kong) error {
		k.registry.RegisterValue(ptr, mapper)
		return nil
	})
}

// NamedMapper registers a mapper to a name.
func NamedMapper(name string, mapper Mapper) Option {
	return OptionFunc(func(k *Kong) error {
		k.registry.RegisterName(name, mapper)
		return nil
	})
}

// Writers overrides the default writers. Useful for testing or interactive use.
func Writers(stdout, stderr io.Writer) Option {
	return OptionFunc(func(k *Kong) error {
		k.Stdout = stdout
		k.Stderr = stderr
		return nil
	})
}

// Bind binds values for hooks and Run() function arguments.
//
// Any arguments passed will be available to the receiving hook functions, but may be omitted. Additionally, *Kong and
// the current *Context will also be made available.
//
// There are two hook points:
//
// 		BeforeApply(...) error
//   	AfterApply(...) error
//
// Called before validation/assignment, and immediately after validation/assignment, respectively.
func Bind(args ...interface{}) Option {
	return OptionFunc(func(k *Kong) error {
		k.bindings.add(args...)
		return nil
	})
}

// BindTo allows binding of implementations to interfaces.
//
// 		BindTo(impl, (*iface)(nil))
func BindTo(impl, iface interface{}) Option {
	return OptionFunc(func(k *Kong) error {
		k.bindings[reflect.TypeOf(iface).Elem()] = reflect.ValueOf(impl)
		return nil
	})
}

// Help printer to use.
func Help(help HelpPrinter) Option {
	return OptionFunc(func(k *Kong) error {
		k.help = help
		return nil
	})
}

// ConfigureHelp sets the HelpOptions to use for printing help.
func ConfigureHelp(options HelpOptions) Option {
	return OptionFunc(func(k *Kong) error {
		k.helpOptions = options
		return nil
	})
}

// UsageOnError configures Kong to display context-sensitive usage if FatalIfErrorf is called with an error.
func UsageOnError() Option {
	return OptionFunc(func(k *Kong) error {
		k.usageOnError = true
		return nil
	})
}

// ClearResolvers clears all existing resolvers.
func ClearResolvers() Option {
	return OptionFunc(func(k *Kong) error {
		k.resolvers = nil
		return nil
	})
}

// Resolvers registers flag resolvers.
func Resolvers(resolvers ...Resolver) Option {
	return OptionFunc(func(k *Kong) error {
		k.resolvers = append(k.resolvers, resolvers...)
		return nil
	})
}

// ConfigurationLoader is a function that builds a resolver from a file.
type ConfigurationLoader func(r io.Reader) (Resolver, error)

// Configuration provides Kong with support for loading defaults from a set of configuration files.
//
// Paths will be opened in order, and "loader" will be used to provide a Resolver which is registered with Kong.
//
// Note: The JSON function is a ConfigurationLoader.
//
// ~ expansion will occur on the provided paths.
func Configuration(loader ConfigurationLoader, paths ...string) Option {
	return OptionFunc(func(k *Kong) error {
		k.loader = loader
		for _, path := range paths {
			resolver, _ := k.LoadConfig(path)
			if resolver != nil {
				k.resolvers = append(k.resolvers, resolver)
			}
		}
		return nil
	})
}

// ExpandPath is a helper function to expand a relative or home-relative path to an absolute path.
//
// eg. ~/.someconf -> /home/alec/.someconf
func ExpandPath(path string) string {
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
