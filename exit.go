package kong

import "errors"

const (
	exitOk    = 0
	exitNotOk = 1

	// Semantic exit codes from https://github.com/square/exit?tab=readme-ov-file#about
	exitUsageError = 80
)

// ExitCoder is an interface that may be implemented by an error value to
// provide an integer exit code. The method ExitCode should return an integer
// that is intended to be used as the exit code for the application.
type ExitCoder interface {
	ExitCode() int
}

// exitCodeFromError returns the exit code for the given error.
// If err implements the exitCoder interface, the ExitCode method is called.
// Otherwise, exitCodeFromError returns 0 if err is nil, and 1 if it is not.
func exitCodeFromError(err error) int {
	var e ExitCoder
	if errors.As(err, &e) {
		return e.ExitCode()
	} else if err == nil {
		return exitOk
	}

	return exitNotOk
}

type exitCodeError struct {
	code int
	err  error
}

func (e *exitCodeError) Error() string { return e.err.Error() }
func (e *exitCodeError) Unwrap() error { return e.err }
func (e *exitCodeError) ExitCode() int { return e.code }

func exitAsUsageError(err error) error {
	return &exitCodeError{code: exitUsageError, err: err}
}
