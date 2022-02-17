package kong

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
)

// ParseError is the error type returned by Kong.Parse().
//
// It contains the parse Context that triggered the error.
type ParseError struct {
	error
	Context *Context
}

// Unwrap returns the original cause of the error.
func (p *ParseError) Unwrap() error { return p.error }

// NOTE: The code bellow is taken from https://github.com/pkg/errors. BSD-2 license.

// withStackError is the error that wraps caller stack.
//
// It is a helper type that formats stack trace properly.
type withStackError struct {
	error
	stack []uintptr
}

// newWithStackError creates new error with the call stack
func newWithStackError(err error) *withStackError {
	return &withStackError{
		error: err,
		stack: callers(),
	}
}

// Unwrap returns the original cause of the error.
func (w *withStackError) Unwrap() error { return w.error }

// Format formats the error according to the fmt.Formatter interface
//
//    %s    original error message
//    %q    quoted original error message
//    %v    equivalent to %s
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//    %+v   prints original error with verb %+v and also formats
//          the call stack in format <f>:<n>, where
//
//          <f> — function name and path of source file relative to the compile time
//          GOPATH separated by \n\t (<funcname>\n\t<path>)
//          <n> — source line
func (w *withStackError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v", w.Unwrap())
			for _, pc := range w.stack {
				f := frame(pc)
				_, _ = io.WriteString(s, "\n"+f.name())
				_, _ = io.WriteString(s, "\n\t")
				_, _ = io.WriteString(s, f.file())
				_, _ = io.WriteString(s, ":")
				_, _ = io.WriteString(s, strconv.Itoa(f.line()))
			}
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, w.Error())
	case 'q':
		fmt.Fprintf(s, "%q", w.Error())
	}
}

// frame represents a program counter inside a stack frame.
type frame uintptr

// pc returns the program counter for this frame;
// multiple frames may have the same PC value.
func (f frame) pc() uintptr { return uintptr(f) - 1 }

// file returns the full path to the file that contains the
// function for this frame's pc.
func (f frame) file() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.pc())
	return file
}

// line returns the line number of source code of the
// function for this frame's pc.
func (f frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

// name returns the name of this function, if known.
func (f frame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

func callers() []uintptr {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	return pcs[0:n]
}
