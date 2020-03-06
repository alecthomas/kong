package kong

import (
	"fmt"
	"strings"

	"github.com/posener/complete"
)

// Completer implements a completion method, in which given command
// line arguments returns a list of options it completes.
//
// See https://github.com/posener/complete for details.
type Completer = complete.Predictor

// CompleterArgs describes command line arguments used by a Completer to predict options.
//
// See https://github.com/posener/complete for details.
type CompleterArgs = complete.Args

// A CompleterFunc is a function that implements the Completer interface.
type CompleterFunc func(CompleterArgs) []string

// Predict completion results.
func (p CompleterFunc) Predict(args CompleterArgs) []string {
	if p == nil {
		return nil
	}
	return p(args)
}

// CompleteNothing returns a nil Completer that indicates no prediction is to be made.
func CompleteNothing() Completer { return complete.PredictNothing }

// CompleteAnything returns a Completer that expects something, but nothing particular, such as a number or arbitrary name.
func CompleteAnything() Completer { return complete.PredictAnything }

// CompleteSet returns a Completer that predicts provided options
func CompleteSet(options ...string) Completer { return complete.PredictSet(options...) }

// CompleteOr unions two predicate functions, so that the result predicate returns the union of their predication
func CompleteOr(completers ...Completer) Completer { return complete.PredictOr(completers...) }

// CompleteDirs will search for directories in the given started to be typed path.
//
// If no path was started to be typed, it will complete to directories in the current working directory.
func CompleteDirs(pattern string) Completer { return complete.PredictDirs(pattern) }

// CompleteFiles will search for files matching the given pattern in the started to be typed path.
//
// If no path was started to be typed, it will complete to files that match the pattern in the
// current working directory. To match any file, use "*" as pattern. To match go files use "*.go", and so on.
func CompleteFiles(pattern string) Completer { return complete.PredictFiles(pattern) }

func tagCompleter(tag *Tag, completers map[string]Completer) (Completer, error) {
	if tag == nil || tag.Completer == "" {
		if tag != nil && tag.Type != "" {
			switch tag.Type {
			case "path":
				return CompleteOr(CompleteFiles("*"), CompleteDirs("*")), nil

			case "existingfile":
				return CompleteFiles("*"), nil

			case "existingdir":
				return CompleteDirs("*"), nil
			}
		}
		return nil, nil
	}
	if completers == nil {
		completers = map[string]Completer{}
	}
	completerName := tag.Completer
	completer, ok := completers[completerName]
	if !ok {
		return nil, fmt.Errorf("no completer with name %q", completerName)
	}
	return completer, nil
}

func valueCompleter(value *Value, completers map[string]Completer) (Completer, error) {
	if value == nil {
		return nil, nil
	}
	completer, err := tagCompleter(value.Tag, completers)
	if err != nil {
		return nil, err
	}
	if completer != nil {
		return completer, nil
	}
	switch {
	case value.IsBool():
		return CompleteNothing(), nil

	case value.Enum != "":
		enumVals := make([]string, 0, len(value.EnumMap()))
		for enumVal := range value.EnumMap() {
			enumVals = append(enumVals, enumVal)
		}
		return CompleteSet(enumVals...), nil

	default:
		return CompleteAnything(), nil
	}
}

func positionalCompleters(args []*Positional, completers map[string]Completer) ([]Completer, error) {
	res := make([]Completer, len(args))
	var err error
	for i, arg := range args {
		res[i], err = valueCompleter(arg, completers)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func flagCompleter(flag *Flag, completers map[string]Completer) (Completer, error) {
	return valueCompleter(flag.Value, completers)
}

// positionalCompleter is a completer for positional arguments
type positionalCompleter struct {
	Completers []Completer
	Flags      []*Flag
}

// Predict implements complete.Predict
func (p *positionalCompleter) Predict(args CompleterArgs) []string {
	completer := p.completer(args)
	if completer == nil {
		return []string{}
	}
	return completer.Predict(args)
}

func (p *positionalCompleter) completer(args CompleterArgs) Completer {
	position := p.completerIndex(args)
	if position < 0 || position > len(p.Completers)-1 {
		return nil
	}
	return p.Completers[position]
}

// completerIndex returns the index in completers to use. Returns -1 if no completer should be used.
func (p *positionalCompleter) completerIndex(args CompleterArgs) int {
	idx := 0
	for i := 0; i < len(args.Completed); i++ {
		if !p.nonCompleterPos(args, i) {
			idx++
		}
	}
	return idx
}

// nonCompleterPos returns true if the value at this position is either a flag or a flag's argument
func (p *positionalCompleter) nonCompleterPos(args CompleterArgs, pos int) bool {
	if pos < 0 || pos > len(args.All)-1 {
		return false
	}
	val := args.All[pos]
	if p.valIsFlag(val) {
		return true
	}
	if pos == 0 {
		return false
	}
	prev := args.All[pos-1]
	return p.nextValueIsFlagArg(prev)
}

// valIsFlag returns true if the value matches a flag from the configuration
func (p *positionalCompleter) valIsFlag(val string) bool {
	val = strings.Split(val, "=")[0]

	for _, flag := range p.Flags {
		if flag == nil {
			continue
		}
		if val == "--"+flag.Name {
			return true
		}
		if flag.Short == 0 {
			continue
		}
		if strings.HasPrefix(val, "-"+string(flag.Short)) {
			return true
		}
	}
	return false
}

// nextValueIsFlagArg returns true if the value matches an ArgFlag and doesn't contain an equal sign.
func (p *positionalCompleter) nextValueIsFlagArg(val string) bool {
	if strings.Contains(val, "=") {
		return false
	}
	for _, flag := range p.Flags {
		if flag.IsBool() {
			continue
		}
		if val == "--"+flag.Name {
			return true
		}
		if flag.Short == 0 {
			continue
		}
		if strings.HasPrefix(val, "-"+string(flag.Short)) {
			return true
		}
	}
	return false
}
