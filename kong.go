package kong

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	model, err := build(ast)
	if err != nil {
		return nil, err
	}
	model.Name = filepath.Base(os.Args[0])
	k := &Kong{
		Model:       model,
		terminate:   os.Exit,
		stdout:      os.Stdout,
		stderr:      os.Stderr,
		help:        defaultHelpTemplate,
		helpContext: map[string]interface{}{},
		helpFuncs:   template.FuncMap{},
	}
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
	k.reset(k.Model)
	ctx := &ParseContext{
		Scan: Scan(args...),
	}
	err = ctx.applyNode(k.Model)
	return strings.Join(ctx.Command, " "), err
}

// Recursively reset values to defaults (as specified in the grammar) or the zero value.
func (k *Kong) reset(node *Node) {
	for _, flag := range node.Flags {
		flag.Value.Reset()
	}
	for _, pos := range node.Positional {
		pos.Reset()
	}
	for _, branch := range node.Children {
		if branch.Argument != nil {
			arg := branch.Argument.Argument
			arg.Reset()
			k.reset(&branch.Argument.Node)
		} else {
			k.reset(branch.Command)
		}
	}
}

func (k *Kong) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, k.Model.Name+": "+format, args...)
}

func (k *Kong) FatalIfErrorf(err error, args ...interface{}) {
	if err == nil {
		return
	}
	msg := err.Error()
	if len(args) == 0 {
		msg = fmt.Sprintf(args[0].(string), args...) + ": " + err.Error()
	}
	k.Errorf("%s", msg)
	k.terminate(1)
}
