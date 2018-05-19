package kong

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Error struct{ msg string }

func (e Error) Error() string { return e.msg }

func fail(format string, args ...interface{}) {
	panic(Error{fmt.Sprintf(format, args...)})
}

// Kong is the main parser type.
type Kong struct {
	Model *Application
	// Termination function (defaults to os.Exit)
	terminate func(int)

	stdout io.Writer
	stderr io.Writer

	help        *template.Template
	helpContext map[string]interface{}
	helpFuncs   template.FuncMap
}

// New creates a new Kong parser into ast.
func New(ast interface{}, options ...Option) (*Kong, error) {
	model, err := build(ast)
	if err != nil {
		return nil, err
	}
	model.Name = filepath.Base(os.Args[0])
	k := &Kong{
		Model:       model,
		terminate:   os.Exit,
		stdout:      os.Stdout,
		stderr:      os.Stderr,
		help:        defaultHelpTemplate,
		helpContext: map[string]interface{}{},
		helpFuncs:   template.FuncMap{},
	}
	for _, option := range options {
		option(k)
	}
	return k, nil
}

// Parse arguments into target.
func (k *Kong) Parse(args []string) (command string, err error) {
	defer func() {
		msg := recover()
		if test, ok := msg.(Error); ok {
			err = test
		} else if msg != nil {
			panic(msg)
		}
	}()
	k.reset(k.Model)
	ctx := &ParseContext{
		Scan: Scan(args...),
	}
	err = ctx.applyNode(k.Model)
	return strings.Join(ctx.Command, " "), err
}

// Recursively reset values to defaults (as specified in the grammar) or the zero value.
func (k *Kong) reset(node *Node) {
	for _, flag := range node.Flags {
		flag.Value.Reset()
	}
	for _, pos := range node.Positional {
		pos.Reset()
	}
	for _, branch := range node.Children {
		if branch.Argument != nil {
			arg := branch.Argument.Argument
			arg.Reset()
			k.reset(&branch.Argument.Node)
		} else {
			k.reset(branch.Command)
		}
	}
}

type ParseContext struct {
	Scan    *Scanner
	Command []string
	Flags   []*Flag
}

func (p *ParseContext) applyNode(node *Node) (err error) { // nolint: gocyclo
	positional := 0
	p.Flags = append(p.Flags, node.Flags...)

	for token := p.Scan.Pop(); token.Type != EOLToken; token = p.Scan.Pop() {
		switch token.Type {
		case UntypedToken:
			switch {
			// -- indicates end of parsing. All remaining arguments are treated as positional arguments only.
			case token.Value == "--":
				args := []string{}
				for {
					token = p.Scan.Pop()
					if token.Type == EOLToken {
						break
					}
					args = append(args, token.Value)
				}
				// Note: tokens must be pushed in reverse order.
				for i := range args {
					p.Scan.PushTyped(args[len(args)-1-i], PositionalArgumentToken)
				}

			// Long flag.
			case strings.HasPrefix(token.Value, "--"):
				// Parse it and push the tokens.
				parts := strings.SplitN(token.Value[2:], "=", 2)
				if len(parts) > 1 {
					p.Scan.PushTyped(parts[1], FlagValueToken)
				}
				p.Scan.PushTyped(parts[0], FlagToken)

			// Short flag.
			case strings.HasPrefix(token.Value, "-"):
				// Note: tokens must be pushed in reverse order.
				p.Scan.PushTyped(token.Value[2:], ShortFlagTailToken)
				p.Scan.PushTyped(token.Value[1:2], ShortFlagToken)

			default:
				p.Scan.PushTyped(token.Value, PositionalArgumentToken)
			}

		case ShortFlagTailToken:
			// Note: tokens must be pushed in reverse order.
			p.Scan.PushTyped(token.Value[1:], ShortFlagTailToken)
			p.Scan.PushTyped(token.Value[0:1], ShortFlagToken)

		case FlagToken:
			if err := matchFlags(p.Flags, token, p.Scan, func(f *Flag) bool {
				return f.Name == token.Value
			}); err != nil {
				return err
			}

		case ShortFlagToken:
			if err := matchFlags(p.Flags, token, p.Scan, func(f *Flag) bool {
				return string(f.Name) == token.Value
			}); err != nil {
				return err
			}

		case FlagValueToken:
			return fmt.Errorf("unexpected flag argument %q", token.Value)

		case PositionalArgumentToken:
			p.Scan.PushToken(token)
			// Ensure we've consumed all positional arguments.
			if positional < len(node.Positional) {
				arg := node.Positional[positional]
				err := arg.Decode(p.Scan)
				if err != nil {
					return err
				}
				p.Command = append(p.Command, "<"+arg.Name+">")
				positional++
				break
			}

			// After positional arguments have been consumed, handle commands and branching arguments.
			for _, branch := range node.Children {
				switch {
				case branch.Command != nil:
					if branch.Command.Name == token.Value {
						p.Scan.Pop()
						p.Command = append(p.Command, branch.Command.Name)
						return p.applyNode(branch.Command)
					}

				case branch.Argument != nil:
					arg := branch.Argument.Argument
					if err := arg.Decode(p.Scan); err == nil {
						p.Command = append(p.Command, "<"+arg.Name+">")
						return p.applyNode(&branch.Argument.Node)
					}
				}
			}
			return fmt.Errorf("unexpected positional argument %s", token)

		default:
			return fmt.Errorf("unexpected token %s", token)
		}
	}

	if err := checkMissingPositionals(positional, node.Positional); err != nil {
		return err
	}

	if err := checkMissingChildren(node.Children); err != nil {
		return err
	}

	if err := checkMissingFlags(node.Children, p.Flags); err != nil {
		return err
	}

	return nil
}

func checkMissingFlags(children []*Branch, flags []*Flag) error {
	// Only check required missing fields at the last child.
	if len(children) > 0 {
		return nil
	}
	missing := []string{}
	for _, flag := range flags {
		if !flag.Required || flag.Set {
			continue
		}
		missing = append(missing, flag.Name)
	}
	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf("missing flags: %s", strings.Join(missing, ", "))
}

func checkMissingChildren(children []*Branch) error {
	missing := []string{}
	for _, child := range children {
		if child.Argument != nil {
			if !child.Argument.Argument.Required {
				continue
			}
			missing = append(missing, "<"+child.Argument.Name+">")
		} else {
			missing = append(missing, child.Command.Name)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf("expected one of %s", strings.Join(missing, ", "))
}

// If we're missing any positionals and they're required, return an error.
func checkMissingPositionals(positional int, values []*Value) error {
	// All the positionals are in.
	if positional == len(values) {
		return nil
	}

	// We're low on supplied positionals, but the missing one is optional.
	if !values[positional].Required {
		return nil
	}

	missing := []string{}
	for ; positional < len(values); positional++ {
		missing = append(missing, "<"+values[positional].Name+">")
	}
	return fmt.Errorf("missing positional arguments %s", strings.Join(missing, " "))
}

func matchFlags(flags []*Flag, token Token, scan *Scanner, matcher func(f *Flag) bool) (err error) {
	defer func() {
		msg := recover()
		if test, ok := msg.(Error); ok {
			err = fmt.Errorf("%s %s", token, test)
		} else if msg != nil {
			panic(msg)
		}
	}()
	for _, flag := range flags {
		// Found a matching flag.
		if flag.Name == token.Value {
			err := flag.Decode(scan)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("unknown flag --%s", token.Value)
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
	k.terminate(1)
}
