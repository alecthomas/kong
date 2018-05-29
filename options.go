package kong

import (
	"io"
	"reflect"
)

// Options apply optional changes to the Kong application.
//
// Note that Options are applied twice: once just prior to the grammar is constructed and once after. In the
// former case, Kong.Application will be nil.
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
		if k.Application != nil {
			k.Name = name
		}
	}
}

// Description sets the application description.
func Description(description string) Option {
	return func(k *Kong) {
		if k.Application != nil {
			k.Help = description
		}
	}
}

// Writers overrides the default writers. Useful for testing or interactive use.
func Writers(stdout, stderr io.Writer) Option {
	return func(k *Kong) {
		k.Stdout = stdout
		k.Stderr = stderr
	}
}

// Hook to execute when a command, flag or positional argument is encountered.
func Hook(ptr interface{}, hook HookFunction) Option {
	key := reflect.ValueOf(ptr)
	if key.Kind() != reflect.Ptr {
		panic("expected a pointer")
	}
	return func(k *Kong) {
		k.hooks[key] = hook
	}
}
