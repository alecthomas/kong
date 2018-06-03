package kong

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

type HookFunction func(ctx *Context, path *Path) error

// Error reported by Kong.
type Error struct{ msg string }

func (e Error) Error() string { return e.msg }

func fail(format string, args ...interface{}) {
	panic(Error{fmt.Sprintf(format, args...)})
}

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
	*Application

	// Termination function (defaults to os.Exit)
	Exit func(int)

	Stdout io.Writer
	Stderr io.Writer

	hooks         map[reflect.Value]HookFunction
	noDefaultHelp bool
}

// New creates a new Kong parser on grammar.
//
// See the README (https://github.com/alecthomas/kong) for usage instructions.
func New(grammar interface{}, options ...Option) (*Kong, error) {
	k := &Kong{
		Exit:   os.Exit,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		hooks:  map[reflect.Value]HookFunction{},
	}

	model, err := build(grammar, k.extraFlags())
	if err != nil {
		return k, err
	}
	k.Application = model
	k.Name = filepath.Base(os.Args[0])

	for _, option := range options {
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
	helpFlag := &Flag{
		Value: Value{
			Name:    "help",
			Help:    "Show context-sensitive help.",
			Flag:    true,
			Value:   reflect.ValueOf(&helpValue).Elem(),
			Decoder: kindDecoders[reflect.Bool],
		},
	}
	hook := Hook(&helpValue, Help(defaultHelpTemplate, nil))
	hook(k)
	return []*Flag{helpFlag}
}

// Path parses the command-line, validating and collecting matching grammar nodes.
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
		if hook := k.hooks[key]; hook != nil {
			if err := hook(ctx, trace); err != nil {
				return err
			}
		}
	}
	return nil
}

// Printf writes a message to Kong.Stdout with the application name prefixed.
func (k *Kong) Printf(format string, args ...interface{}) {
	fmt.Fprintf(k.Stdout, k.Name+": "+format, args...)
}

// Errorf writes a message to Kong.Stderr with the application name prefixed.
func (k *Kong) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(k.Stderr, k.Name+": "+format, args...)
}

// FatalIfError terminates with an error message if err != nil.
func (k *Kong) FatalIfErrorf(err error, args ...interface{}) {
	if err == nil {
		return
	}
	msg := err.Error()
	if len(args) > 0 {
		msg = fmt.Sprintf(args[0].(string), args...) + ": " + err.Error()
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
