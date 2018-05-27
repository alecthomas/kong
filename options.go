package kong

import (
	"io"
	"reflect"
	"text/template"
)

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
	return func(k *Kong) { k.Model.Name = name }
}

// Description sets the application description.
func Description(description string) Option {
	return func(k *Kong) { k.Model.Help = description }
}

// HelpTemplate overrides the default help template.
func HelpTemplate(template *template.Template) Option {
	return func(k *Kong) { k.help = template }
}

// HelpContext sets extra context in the help template.
//
// The key "Application" will always be available and is the root of the application model.
func HelpContext(context map[string]interface{}) Option {
	return func(k *Kong) { k.helpContext = context }
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
