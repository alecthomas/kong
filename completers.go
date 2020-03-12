package kong

import (
	"strings"
	"unicode"
)

// Completers contains custom Completers used to generate completion options.
//
// They can be used with an annotation like `completer='myCustomCompleter'` where "myCustomCompleter" is a
// key in Completers.
type Completers map[string]Completer

// Apply completers to Kong as a configuration option.
func (c Completers) Apply(k *Kong) error {
	k.completers = c
	return nil
}

// Completer implements a completion method, in which given command
// line arguments returns a list of options it completes.
//
// See https://github.com/posener/complete for details.
type Completer interface {
	Options(CompleterArgs) []string
}

// CompleterArgs describes command line arguments used by a Completer.
type CompleterArgs []string

// All lists of all arguments in command line (not including the command itself)
func (a CompleterArgs) All() []string {
	if len(a) == 0 {
		return []string{}
	}
	return a[1:]
}

// Completed lists of all completed arguments in command line,
// If the last one is still being typed - no space after it,
// it won't appear in this list of arguments.
func (a CompleterArgs) Completed() []string {
	all := a.All()
	if len(all) > 0 {
		return all[:len(all)-1]
	}
	return []string{}
}

// Last argument in command line, the one being typed, if the last
// character in the command line is a space, this argument will be empty,
// otherwise this would be the last word.
func (a CompleterArgs) Last() string {
	if len(a) == 0 {
		return ""
	}
	return a[len(a)-1]
}

// LastCompleted is the last argument that was fully typed.
// If the last character in the command line is space, this would be the
// last word, otherwise, it would be the word before that.
func (a CompleterArgs) LastCompleted() string {
	comp := a.Completed()
	if len(comp) > 0 {
		return comp[len(comp)-1]
	}
	return ""
}

func newCompleterArgs(line string) CompleterArgs {
	var args CompleterArgs = strings.Fields(line)

	if len(args) == 0 {
		return args
	}

	// Add empty field if the last field was completed.
	if unicode.IsSpace(rune(line[len(line)-1])) {
		args = append(args, "")
	}

	// if the last arg is in the form a=b, split it
	lastArgParts := strings.Split(args.Last(), "=")
	return append(args[:len(args)-1], lastArgParts...)
}

// A CompleterFunc is a function that implements the Completer interface.
type CompleterFunc func(CompleterArgs) []string

// Options completion results.
func (p CompleterFunc) Options(args CompleterArgs) []string {
	if p == nil {
		return nil
	}
	return p(args)
}

// CompleteNothing returns a nil Completer that indicates no completion is to be made.
func CompleteNothing() Completer {
	return nil
}

// CompleteAnything returns a Completer that expects something, but nothing particular, such as a number or arbitrary name.
func CompleteAnything() CompleterFunc {
	return func(_ CompleterArgs) []string {
		return nil
	}
}

// CompleteSet returns a Completer that returns the provided options
func CompleteSet(options ...string) CompleterFunc {
	return func(_ CompleterArgs) []string {
		return options
	}
}

// CompleteOr unions two predicate functions, so that the result predicate returns the union of their predication
func CompleteOr(completers ...Completer) Completer {
	return CompleterFunc(func(args CompleterArgs) []string {
		result := make([]string, 0, len(completers))
		for _, c := range completers {
			if c != nil {
				result = append(result, c.Options(args)...)
			}
		}
		return result
	})
}
