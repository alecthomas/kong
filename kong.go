package kong

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

// Error reported by Kong.
type Error struct{ msg string }

func (e Error) Error() string { return e.msg }

func fail(format string, args ...interface{}) {
	panic(Error{fmt.Sprintf(format, args...)})
}

// Must creates a new Parser or panics if there is an error.
func Must(ast interface{}, options ...Option) *Kong {
	k, err := New(ast, options...)
	if err != nil {
		panic(err)
	}
	return k
}

// Kong is the main parser type.
type Kong struct {
	// Grammar model.
	Model *Application

	// Termination function (defaults to os.Exit)
	Exit func(int)

	Stdout io.Writer
	Stderr io.Writer

	before        map[reflect.Value]HookFunc
	resolvers     []ResolverFunc
	registry      *Registry
	noDefaultHelp bool
	help          func(*Context) error

	// Set temporarily by Options. These are applied after build().
	postBuildOptions []Option
}

// New creates a new Kong parser on grammar.
//
// See the README (https://github.com/alecthomas/kong) for usage instructions.
func New(grammar interface{}, options ...Option) (*Kong, error) {
	k := &Kong{
		Exit:     os.Exit,
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		before:   map[reflect.Value]HookFunc{},
		registry: NewRegistry().RegisterDefaults(),
		help:     PrintHelp,
	}

	for _, option := range options {
		option(k)
	}

	model, err := build(k, grammar)
	if err != nil {
		return k, err
	}
	model.Name = filepath.Base(os.Args[0])
	k.Model = model

	for _, option := range k.postBuildOptions {
		option(k)
	}

	return k, nil
}

// Provide additional builtin flags, if any.
func (k *Kong) extraFlags() []*Flag {
	if k.noDefaultHelp {
		return nil
	}
	helpValue := false
	value := reflect.ValueOf(&helpValue).Elem()
	helpFlag := &Flag{
		Value: &Value{
			Name:   "help",
			Help:   "Show context-sensitive help.",
			Value:  value,
			Tag:    &Tag{},
			Mapper: k.registry.ForValue(value),
		},
	}
	helpFlag.Flag = helpFlag
	hook := Hook(&helpValue, func(ctx *Context, path *Path) error {
		err := PrintHelp(ctx)
		if err != nil {
			return err
		}
		k.Exit(1)
		return nil
	})
	hook(k)
	return []*Flag{helpFlag}
}

// Trace parses the command-line, validating and collecting matching grammar nodes.
func (k *Kong) Trace(args []string) (*Context, error) {
	return Trace(k, args)
}

// Parse arguments into target.
//
// The returned "command" is a space separated path to the final selected command, if any. Commands appear as
// the command name while positional arguments are the argument name surrounded by "<argument>".
func (k *Kong) Parse(args []string) (command string, err error) {
	defer catch(&err)
	ctx, err := k.Trace(args)
	if err != nil {
		return "", err
	}
	if err := k.applyHooks(ctx); err != nil {
		return "", err
	}
	if ctx.Error != nil {
		return "", ctx.Error
	}
	if err = ctx.Validate(); err != nil {
		return "", err
	}
	return ctx.Apply()
}

func (k *Kong) applyHooks(ctx *Context) error {
	for _, trace := range ctx.Path {
		var key reflect.Value
		switch {
		case trace.App != nil:
			key = trace.App.Target
		case trace.Argument != nil:
			key = trace.Argument.Target
		case trace.Command != nil:
			key = trace.Command.Target
		case trace.Positional != nil:
			key = trace.Positional.Value
		case trace.Flag != nil:
			key = trace.Flag.Value.Value
		default:
			panic("unsupported Path")
		}
		if key.IsValid() {
			key = key.Addr()
		}
		if hook := k.before[key]; hook != nil {
			if err := hook(ctx, trace); err != nil {
				return err
			}
		}
	}
	return nil
}

// Printf writes a message to Kong.Stdout with the application name prefixed.
func (k *Kong) Printf(format string, args ...interface{}) {
	fmt.Fprintf(k.Stdout, k.Model.Name+": "+format, args...)
}

// Errorf writes a message to Kong.Stderr with the application name prefixed.
func (k *Kong) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(k.Stderr, k.Model.Name+": error: "+format, args...)
}

// FatalIfErrorf terminates with an error message if err != nil.
func (k *Kong) FatalIfErrorf(err error, args ...interface{}) {
	if err == nil {
		return
	}
	msg := err.Error()
	if len(args) > 0 {
		msg = fmt.Sprintf(args[0].(string), args[1:]...) + ": " + err.Error()
	}
	k.Errorf("%s\n", msg)
	k.Exit(1)
}

func catch(err *error) {
	msg := recover()
	if test, ok := msg.(Error); ok {
		*err = test
	} else if msg != nil {
		panic(msg)
	}
}
