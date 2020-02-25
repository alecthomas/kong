package kong

import (
	"fmt"
	"strings"

	"github.com/posener/complete"
)

// Predictor implements a predict method, in which given command line arguments returns a list of options it predicts.
// Taken from github.com/posener/complete
type Predictor interface {
	Predict(PredictorArgs) []string
}

// PredictorArgs describes command line arguments used by a Predictor to predict options.
// Taken from Args in github.com/posener/complete
type PredictorArgs struct {
	// All lists of all arguments in command line (not including the command itself)
	All []string
	// Completed lists of all completed arguments in command line,
	// If the last one is still being typed - no space after it,
	// it won't appear in this list of arguments.
	Completed []string
	// Last argument in command line, the one being typed, if the last
	// character in the command line is a space, this argument will be empty,
	// otherwise this would be the last word.
	Last string
	// LastCompleted is the last argument that was fully typed.
	// If the last character in the command line is space, this would be the
	// last word, otherwise, it would be the word before that.
	LastCompleted string
}

// NewPredictor returns a Predictor that runs the provided function.
func NewPredictor(fn func(PredictorArgs) []string) Predictor {
	return predictorFunc(fn)
}

type predictorFunc func(PredictorArgs) []string

func (p predictorFunc) Predict(args PredictorArgs) []string {
	if p == nil {
		return nil
	}
	return p(args)
}

// PredictNothing returns a nil Predictor that indicates no prediction is to be made.
func PredictNothing() Predictor {
	return nil
}

// PredictAnything returns a Predictor that expects something, but nothing particular, such as a number or arbitrary name.
func PredictAnything() Predictor {
	return NewPredictor(func(PredictorArgs) []string {
		return nil
	})
}

//PredictSet returns a Predictor that predicts provided options
func PredictSet(options ...string) Predictor {
	return predictorFunc(func(args PredictorArgs) []string {
		return options
	})
}

// newCompletePredictor returns a completePredictor or nil.
// this is needed because nil predictor's have special meaning to complete.Complete
func newCompletePredictor(predictor Predictor) complete.Predictor {
	if predictor == nil {
		return nil
	}
	return &completePredictor{
		predictor: predictor,
	}
}

// completePredictor wraps a Predictor to make it a complete.Predictor
type completePredictor struct {
	predictor Predictor
}

func (c *completePredictor) Predict(args complete.Args) []string {
	return c.predictor.Predict(PredictorArgs{
		All:           args.All,
		Completed:     args.Completed,
		Last:          args.Last,
		LastCompleted: args.LastCompleted,
	})
}

func tagPredictor(tag *Tag, predictors map[string]Predictor) (Predictor, error) {
	if tag == nil || tag.Predictor == "" {
		return nil, nil
	}
	if predictors == nil {
		predictors = map[string]Predictor{}
	}
	predictorName := tag.Predictor
	predictor, ok := predictors[predictorName]
	if !ok {
		return nil, fmt.Errorf("no predictor with name %q", predictorName)
	}
	return predictor, nil
}

func valuePredictor(value *Value, predictors map[string]Predictor) (Predictor, error) {
	if value == nil {
		return nil, nil
	}
	predictor, err := tagPredictor(value.Tag, predictors)
	if err != nil {
		return nil, err
	}
	if predictor != nil {
		return predictor, nil
	}
	switch {
	case value.IsBool():
		return PredictNothing(), nil
	case value.Enum != "":
		enumVals := make([]string, 0, len(value.EnumMap()))
		for enumVal := range value.EnumMap() {
			enumVals = append(enumVals, enumVal)
		}
		return PredictSet(enumVals...), nil
	default:
		return PredictAnything(), nil
	}
}

func positionalPredictors(args []*Positional, predictors map[string]Predictor) ([]Predictor, error) {
	res := make([]Predictor, len(args))
	var err error
	for i, arg := range args {
		res[i], err = valuePredictor(arg, predictors)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func flagPredictor(flag *Flag, predictors map[string]Predictor) (Predictor, error) {
	return valuePredictor(flag.Value, predictors)
}

//positionalPredictor is a predictor for positional arguments
type positionalPredictor struct {
	Predictors []Predictor
	Flags      []*Flag
}

//Predict implements complete.Predict
func (p *positionalPredictor) Predict(args PredictorArgs) []string {
	predictor := p.predictor(args)
	if predictor == nil {
		return []string{}
	}
	return predictor.Predict(args)
}

func (p *positionalPredictor) predictor(args PredictorArgs) Predictor {
	position := p.predictorIndex(args)
	if position < 0 || position > len(p.Predictors)-1 {
		return nil
	}
	return p.Predictors[position]
}

//predictorIndex returns the index in predictors to use. Returns -1 if no predictor should be used.
func (p *positionalPredictor) predictorIndex(args PredictorArgs) int {
	idx := 0
	for i := 0; i < len(args.Completed); i++ {
		if !p.nonPredictorPos(args, i) {
			idx++
		}
	}
	return idx
}

//nonPredictorPos returns true if the value at this position is either a flag or a flag's argument
func (p *positionalPredictor) nonPredictorPos(args PredictorArgs, pos int) bool {
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

//valIsFlag returns true if the value matches a flag from the configuration
func (p *positionalPredictor) valIsFlag(val string) bool {
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

//nextValueIsFlagArg returns true if the value matches an ArgFlag and doesn't contain an equal sign.
func (p *positionalPredictor) nextValueIsFlagArg(val string) bool {
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
