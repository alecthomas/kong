package kong

import (
	"fmt"
	"os"
	"path/filepath"
)

type Kong struct {
	Model *ApplicationModel
	// Termination function (defaults to os.Exit)
	Terminate func(int)
}

// New creates a new Kong parser into grammar.
func New(name, description string, grammar interface{}) (*Kong, error) {
	if name == "" {
		name = filepath.Base(os.Args[0])
	}
	model := &ApplicationModel{
		Name:        name,
		Description: description,
	}
	return &Kong{
		Model:     model,
		Terminate: os.Exit,
	}, nil
}

// Parse arguments into target.
func (k *Kong) Parse(args []string) error {
	return nil
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
	k.Terminate(1)
}
