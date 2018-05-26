package kong

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"text/template"
)

type Hook func(app *Kong, ctx *Context, trace *Trace) error

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
	Model *Application
	// Termination function (defaults to os.Exit)
	Exit func(int)

	Stdout io.Writer
	Stderr io.Writer

	help        *template.Template
	helpContext map[string]interface{}
	helpFuncs   template.FuncMap
	hooks       map[reflect.Value]Hook
}

// New creates a new Kong parser into ast.
func New(ast interface{}, options ...Option) (*Kong, error) {
	k := &Kong{
		Exit:        os.Exit,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		help:        defaultHelpTemplate,
		helpContext: map[string]interface{}{},
		helpFuncs:   template.FuncMap{},
		hooks:       map[reflect.Value]Hook{},
	}

	model, err := build(ast)
	if err != nil {
		return k, err
	}
	k.Model = model
	k.Model.Name = filepath.Base(os.Args[0])

	for _, option := range options {
		option(k)
	}

	return k, nil
}

// Trace parses the command-line, validating and collecting matching grammar nodes.
func (k *Kong) Trace(args []string) (*Context, error) {
	p := &Context{
		app:  k.Model,
		args: args,
		Trace: []*Trace{
			{App: k.Model, Flags: append([]*Flag{}, k.Model.Flags...), Value: k.Model.Target},
		},
	}
	err := p.reset(&p.app.Node)
	if err != nil {
		return nil, err
	}
	p.Error = p.trace(&p.app.Node)
	if err = checkMissingFlags(p.Flags()); err != nil {
		return nil, err
	}
	return p, nil
}

// Hook to execute when a command is encountered.
func (k *Kong) Hook(ptr interface{}, hook Hook) *Kong {
	k.hooks[reflect.ValueOf(ptr)] = hook
	return k
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
	return ctx.Apply()
}

func (k *Kong) applyHooks(ctx *Context) error {
	for _, trace := range ctx.Trace {
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
			panic("unsupported Trace")
		}
		if key.IsValid() {
			key = key.Addr()
		}
		if hook := k.hooks[key]; hook != nil {
			if err := hook(k, ctx, trace); err != nil {
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
	fmt.Fprintf(k.Stderr, k.Model.Name+": "+format, args...)
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
