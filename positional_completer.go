package kong

import (
	"strings"
)

// positionalCompleter is a completer for positional arguments
type positionalCompleter struct {
	Completers []Completer
	Flags      []*Flag
}

func (p *positionalCompleter) Options(args CompleterArgs) []string {
	completer := p.completer(args)
	if completer == nil {
		return []string{}
	}
	return completer.Options(args)
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
	completedArgs := args.Completed()
	for i := 0; i < len(completedArgs); i++ {
		if !p.nonCompleterPos(args, i) {
			idx++
		}
	}
	return idx
}

// nonCompleterPos returns true if the value at this position is either a flag or a flag's argument
func (p *positionalCompleter) nonCompleterPos(args CompleterArgs, pos int) bool {
	allArgs := args.All()
	if pos < 0 || pos > len(allArgs)-1 {
		return false
	}
	val := allArgs[pos]
	if p.valIsFlag(val) {
		return true
	}
	if pos == 0 {
		return false
	}
	prev := allArgs[pos-1]
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
