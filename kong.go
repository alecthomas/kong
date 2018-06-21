package kong

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// Error reported by Kong.
type Error struct{ msg string }

func (e Error) Error() string { return e.msg }

func fail(format string, args ...interface{}) {
	panic(Error{msg: fmt.Sprintf(format, args...)})
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
	usageOnError  bool
	help          HelpPrinter
	helpOptions   HelpPrinterOptions

	// Set temporarily by Options. These are applied after build().
	postBuildOptions []Option
}

// New creates a new Kong parser on grammar.
//
// See the README (https://github.com/alecthomas/kong) for usage instructions.
func New(grammar interface{}, options ...Option) (*Kong, error) {
	k := &Kong{
		Exit:      os.Exit,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
		before:    map[reflect.Value]HookFunc{},
		registry:  NewRegistry().RegisterDefaults(),
		resolvers: []ResolverFunc{Envars()},
	}

	for _, option := range options {
		if err := option(k); err != nil {
			return nil, err
		}
	}

	if k.help == nil {
		k.help = DefaultHelpPrinter
	}

	model, err := build(k, grammar)
	if err != nil {
		return k, err
	}
	model.Name = filepath.Base(os.Args[0])
	k.Model = model

	for _, option := range k.postBuildOptions {
		if err := option(k); err != nil {
			return nil, err
		}
	}
	k.postBuildOptions = nil

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
			Target: value,
			Tag:    &Tag{},
			Mapper: k.registry.ForValue(value),
		},
	}
	helpFlag.Flag = helpFlag
	hook := Hook(&helpValue, func(ctx *Context, path *Path) error {
		options := k.helpOptions
		options.Summary = false
		err := k.help(options, ctx)
		if err != nil {
			return err
		}
		k.Exit(1)
		return nil
	})
	_ = hook(k)
	return []*Flag{helpFlag}
}

// Help writes help for the given error to the stdout io.Writer associated with this Kong.
//
// "err" should be the error returned by Parse().
//
// See Help() and Writers() for overriding the help function and stdout, respectively.
func (k *Kong) Help(err error) error {
	var ctx *Context
	if perr, ok := err.(*ParseError); ok {
		ctx = perr.Context
	} else {
		ctx, err = Trace(k, nil)
		if err != nil {
			return err
		}
	}
	return k.help(k.helpOptions, ctx)
}

// Parse arguments into target.
//
// The returned "command" is a space separated path to the final selected command, if any. Commands appear as
// the command name while positional arguments are the argument name surrounded by "<argument>".
//
// Will return a ParseError if a *semantically* invalid command-line is encountered (as opposed to a syntactically
// invalid one, which will report a normal error).
func (k *Kong) Parse(args []string) (command string, err error) {
	defer catch(&err)
	ctx, err := Trace(k, args)
	if err != nil {
		return "", err
	}
	if err = k.applyHooks(ctx); err != nil {
		return "", &ParseError{error: err, Context: ctx}
	}
	if ctx.Error != nil {
		return "", &ParseError{error: ctx.Error, Context: ctx}
	}
	if err = ctx.Validate(); err != nil {
		return "", &ParseError{error: err, Context: ctx}
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
			key = trace.Positional.Target
		case trace.Flag != nil:
			key = trace.Flag.Value.Target
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

func formatMultilineMessage(w io.Writer, leader string, format string, args ...interface{}) {
	lines := strings.Split(fmt.Sprintf(format, args...), "\n")
	fmt.Fprintf(w, "%s%s\n", leader, lines[0])
	for _, line := range lines[1:] {
		fmt.Fprintf(w, "%*s%s\n", len(leader), " ", line)
	}
}

// Printf writes a message to Kong.Stdout with the application name prefixed.
func (k *Kong) Printf(format string, args ...interface{}) *Kong {
	formatMultilineMessage(k.Stdout, k.Model.Name+": ", format, args...)
	return k
}

// Errorf writes a message to Kong.Stderr with the application name prefixed.
func (k *Kong) Errorf(format string, args ...interface{}) *Kong {
	formatMultilineMessage(k.Stderr, k.Model.Name+": error: ", format, args...)
	return k
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
	k.Errorf("%s", msg)
	// Maybe display usage information.
	if err, ok := err.(*ParseError); ok && k.usageOnError {
		fmt.Fprintln(k.Stdout)
		options := k.helpOptions
		options.Summary = true
		_ = k.help(options, err.Context)
	}
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
