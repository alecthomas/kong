package kong

import (
	"fmt"
	"strings"
)

// ParseError is the error type returned by Kong.Parse().
//
// It contains the parse Context that triggered the error.
type ParseError struct {
	error
	Context *Context
}

// Cause returns the original cause of the error.
func (p *ParseError) Cause() error { return p.error }

type missingChildError struct {
	missing []string
}

func (m *missingChildError) Error() string {
	if len(m.missing) > 5 {
		m.missing = append(m.missing[:5], "...")
	}
	if len(m.missing) == 1 {
		return fmt.Sprintf("expected %s", m.missing[0])
	}
	return fmt.Sprintf("expected one of %s", strings.Join(m.missing, ",  "))
}

func newMissingChildError(missing []string) *missingChildError {
	return &missingChildError{missing}
}

func isMissingChildError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*missingChildError)
	return ok
}
