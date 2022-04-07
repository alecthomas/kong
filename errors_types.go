// Package kong
package kong

import "fmt"

type (
	// Error A singleton object with a list of errors that can be compared by anchor through ==
	Error struct{}

	// The internal structure of the error object with arguments, template, anchor and error interface
	err struct {
		tpl    string        // Template of error
		args   []interface{} // Other error arguments
		anchor error         // Fixed address constant of error
		errFn  func() string // Function of interface error
	}

	// Err an Error interface
	Err interface {
		Anchor() error       // Anchor by which you can compare two errors with each other
		Error() string       // Error message or error message template
		Args() []interface{} // Original arguments for error template
	}
)

// Anchor Implementation of interface Err
func (err err) Anchor() error { return err.anchor }

// Error Implementation of interface Err
func (err err) Error() string { return err.errFn() }

// Args Implementation of interface Err
func (err err) Args() []interface{} { return err.args }

// Errors Collection of package errors that can be compared with each other
func Errors() *Error { return errSingleton }

// Object constructor
func newErr(obj *err, arg ...interface{}) Err {
	return &err{
		anchor: obj,
		args:   arg,
		errFn:  func() string { return fmt.Sprintf(obj.tpl, arg...) },
	}
}
