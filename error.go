package kong

// ParseError is the error type returned by Kong.Parse().
//
// It contains the parse Context that triggered the error.
type ParseError struct {
	error
	Context *Context
}

// Unwrap returns the original cause of the error.
func (p *ParseError) Unwrap() error { return p.error }

// IsErrInterfaceImplements returns the original cause of the error if type is implements Err interface.
func (p *ParseError) IsErrInterfaceImplements() (err Err, ok bool) { err, ok = p.error.(Err); return }

// MustErr returns the original cause of the error type Err interface anyway
func (p *ParseError) MustErr() (err Err) {
	var ok bool

	if err, ok = p.IsErrInterfaceImplements(); !ok {
		err = Errors().UnknownError(p.Unwrap())
	}

	return
}
