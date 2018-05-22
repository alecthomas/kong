package kong

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

type Error struct{ msg string }

func (e Error) Error() string { return e.msg }

func fail(format string, args ...interface{}) {
	panic(Error{fmt.Sprintf(format, args...)})
}

// Kong is the main parser type.
type Kong struct {
	Model *Application
	// Termination function (defaults to os.Exit)
	terminate func(int)

	stdout io.Writer
	stderr io.Writer

	help        *template.Template
	helpContext map[string]interface{}
	helpFuncs   template.FuncMap
}

// New creates a new Kong parser into ast.
func New(ast interface{}, options ...Option) (*Kong, error) {
	k := &Kong{
		terminate:   os.Exit,
		stdout:      os.Stdout,
		stderr:      os.Stderr,
		help:        defaultHelpTemplate,
		helpContext: map[string]interface{}{},
		helpFuncs:   template.FuncMap{},
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

// Parse arguments into target.
func (k *Kong) Parse(args []string) (command string, err error) {
	defer func() {
		msg := recover()
		if test, ok := msg.(Error); ok {
			err = test
		} else if msg != nil {
			panic(msg)
		}
	}()
	ctx, err := Trace(args, k.Model)
	if err != nil {
		return "", err
	}
	if value := ctx.FlagValue(k.Model.HelpFlag); value.IsValid() && value.Bool() {
		return "", nil
	}
	return ctx.Apply()
}

func (k *Kong) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, k.Model.Name+": "+format, args...)
}

func (k *Kong) FatalIfErrorf(err error, args ...interface{}) {
	if err == nil {
		return
	}
	msg := err.Error()
	if len(args) > 0 {
		msg = fmt.Sprintf(args[0].(string), args...) + ": " + err.Error()
	}
	k.Errorf("%s\n", msg)
	k.terminate(1)
}
