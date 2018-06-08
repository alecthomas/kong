package kong

import (
	"io"
	"reflect"
)

// Options apply optional changes to the Kong application.
type Option func(k *Kong)

// ExitFunction overrides the function used to terminate. This is useful for testing or interactive use.
func ExitFunction(exit func(int)) Option {
	return func(k *Kong) { k.Exit = exit }
}

// NoDefaultHelp disables the default help flags.
func NoDefaultHelp() Option {
	return func(k *Kong) {
		k.noDefaultHelp = true
	}
}

// Name overrides the application name.
func Name(name string) Option {
	return func(k *Kong) {
		k.postBuildOptions = append(k.postBuildOptions, func(k *Kong) {
			k.Model.Name = name
		})
	}
}

// Description sets the application description.
func Description(description string) Option {
	return func(k *Kong) {
		k.postBuildOptions = append(k.postBuildOptions, func(k *Kong) {
			k.Model.Help = description
		})
	}
}

// TypeMapper registers a mapper to a type.
func TypeMapper(typ reflect.Type, mapper Mapper) Option {
	return func(k *Kong) { k.registry.RegisterType(typ, mapper) }
}

// KindMapper registers a mapper to a kind.
func KindMapper(kind reflect.Kind, mapper Mapper) Option {
	return func(k *Kong) { k.registry.RegisterKind(kind, mapper) }
}

// ValueMapper registers a mapper to a field value.
func ValueMapper(ptr interface{}, mapper Mapper) Option {
	return func(k *Kong) { k.registry.RegisterValue(ptr, mapper) }
}

// NamedMapper registers a mapper to a name.
func NamedMapper(name string, mapper Mapper) Option {
	return func(k *Kong) { k.registry.RegisterName(name, mapper) }
}

// Writers overrides the default writers. Useful for testing or interactive use.
func Writers(stdout, stderr io.Writer) Option {
	return func(k *Kong) {
		k.Stdout = stdout
		k.Stderr = stderr
	}
}

// HookFunc is a callback tied to a field of the grammar, called before a value is applied.
type HookFunc func(ctx *Context, path *Path) error

// Hook to aply before a command, flag or positional argument is encountered.
//
// "ptr" is a pointer to a field of the grammar.
func Hook(ptr interface{}, hook HookFunc) Option {
	key := reflect.ValueOf(ptr)
	if key.Kind() != reflect.Ptr {
		panic("expected a pointer")
	}
	return func(k *Kong) {
		k.before[key] = hook
	}
}

type HelpFunction func(*Context) error

// Help function to use.
//
// Defaults to PrintHelp.
func Help(help func(*Context) error) Option {
	return func(k *Kong) {
		k.help = help
	}
}
