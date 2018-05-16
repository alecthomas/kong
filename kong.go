package kong

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type Kong struct {
	Model *Application
	// Termination function (defaults to os.Exit)
	Terminate func(int)
}

// New creates a new Kong parser into ast.
func New(name, description string, ast interface{}) (*Kong, error) {
	if name == "" {
		name = filepath.Base(os.Args[0])
	}
	model, err := build(ast)
	if err != nil {
		return nil, err
	}
	model.Name = name
	model.Help = description
	return &Kong{
		Model:     model,
		Terminate: os.Exit,
	}, nil
}

// Parse arguments into target.
func (k *Kong) Parse(args []string) (command string, err error) {
	defer func() {
		msg := recover()
		if test, ok := msg.(TokenAssertionError); ok {
			err = test
		} else if msg != nil {
			panic(msg)
		}
	}()
	k.reset(k.Model)
	cmd, err := k.applyNode(Scan(args...), k.Model)
	return strings.Join(cmd, " "), err
}

// Recursively reset values to defaults (as specified in the grammar) or the zero value.
func (k *Kong) reset(node *Node) {
	for _, flag := range node.Flags {
		if flag.Default != "" {
			flag.Decoder.Decode(Scan(flag.Default), flag.Value.Value)
		} else {
			flag.Value.Value.Set(reflect.Zero(flag.Value.Value.Type()))
		}
	}
	for _, pos := range node.Positional {
		pos.Value.Set(reflect.Zero(pos.Value.Type()))
	}
	for _, branch := range node.Children {
		if branch.Argument != nil {
			arg := branch.Argument.Argument
			arg.Value.Set(reflect.Zero(arg.Value.Type()))
			k.reset(&branch.Argument.Node)
		} else {
			k.reset(branch.Command)
		}
	}
}

func (k *Kong) applyNode(scan *Scanner, node *Node) (command []string, err error) {
	for token := scan.Pop(); token.Type != EOLToken; token = scan.Pop() {
		switch token.Type {
		case UntypedToken:
			switch {
			// -- indicates end of parsing. All remaining arguments are treated as positional arguments only.
			case token.Value == "--":
				for {
					token = scan.Pop()
					if token.Type == EOLToken {
						break
					}
					scan.PushTyped(token.Value, PositionalArgumentToken)
				}

			// Long flag.
			case strings.HasPrefix(token.Value, "--"):
				// Parse it and push the tokens.
				parts := strings.SplitN(token.Value[2:], "=", 2)
				scan.PushTyped(parts[0], FlagToken)
				if len(parts) > 1 {
					scan.PushTyped(parts[1], FlagValueToken)
				}

				// Short flag.
			case strings.HasPrefix(token.Value, "-"):
				scan.PushTyped(token.Value[1:2], ShortFlagToken)
				scan.PushTyped(token.Value[2:], ShortFlagTailToken)

			default:
				scan.PushTyped(token.Value, PositionalArgumentToken)
			}

		case ShortFlagTailToken:
			scan.PushTyped(token.Value[0:1], ShortFlagToken)
			scan.PushTyped(token.Value[1:], ShortFlagTailToken)

		case FlagToken:
			if err := matchFlags(node.Flags, token, scan, func(f *Flag) bool {
				return f.Name == token.Value
			}); err != nil {
				return nil, err
			}

		case ShortFlagToken:
			if err := matchFlags(node.Flags, token, scan, func(f *Flag) bool {
				return string(f.Name) == token.Value
			}); err != nil {
				return nil, err
			}

		case FlagValueToken:
			return nil, fmt.Errorf("unexpected flag argument %q", token.Value)

		case PositionalArgumentToken:
			scan.PushToken(token)
			for _, branch := range node.Children {
				switch {
				case branch.Command != nil:
					if branch.Command.Name == token.Value {
						scan.Pop()
						command = append(command, branch.Command.Name)
						cmd, err := k.applyNode(scan, branch.Command)
						if err != nil {
							return nil, err
						}
						return append(command, cmd...), nil
					}

				case branch.Argument != nil:
					arg := branch.Argument.Argument
					if err := arg.Decoder.Decode(scan, arg.Value); err == nil {
						command = append(command, "<"+arg.Name+">")
						cmd, err := k.applyNode(scan, &branch.Argument.Node)
						if err != nil {
							return nil, err
						}
						return append(command, cmd...), nil
					}
				}
			}
			return nil, fmt.Errorf("unexpected positional argument %s", token)

		default:
			return nil, fmt.Errorf("unexpected token %s", token)
		}
	}
	return
}

func matchFlags(flags []*Flag, token Token, scan *Scanner, matcher func(f *Flag) bool) (err error) {
	defer func() {
		msg := recover()
		if test, ok := msg.(TokenAssertionError); ok {
			err = fmt.Errorf("%s %s", token, test)
		} else if msg != nil {
			panic(msg)
		}
	}()
	for _, flag := range flags {
		// Found a matching flag.
		if flag.Name == token.Value {
			err := flag.Decoder.Decode(scan, flag.Value.Value)
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
	k.Terminate(1)
}
