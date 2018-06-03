package kong

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Path records the nodes and parsed values from the current command-line.
type Path struct {
	Parent *Node

	// One of these will be non-nil.
	App        *Application
	Positional *Positional
	Flag       *Flag
	Argument   *Argument
	Command    *Command

	// Flags added by this node.
	Flags []*Flag

	// Parsed value for non-commands.
	Value reflect.Value
}

type Context struct {
	App   *Kong
	Path  []*Path // A trace through parsed nodes.
	Error error   // Error that occurred during trace, if any.

	args []string
	scan *Scanner
}

// Selected command or argument.
func (c *Context) Selected() *Node {
	var selected *Node
	for _, path := range c.Path {
		switch {
		case path.Command != nil:
			selected = path.Command
		case path.Argument != nil:
			selected = path.Argument
		}
	}
	return selected
}

// Trace path of "args" through the gammar tree.
//
// The returned Context will include a Path of all commands, arguments, positionals and flags.
func Trace(k *Kong, args []string) (*Context, error) {
	c := &Context{
		App:  k,
		args: args,
		Path: []*Path{
			{App: k.Model, Flags: k.Model.Flags, Value: k.Model.Target},
		},
	}
	err := c.reset(&c.App.Model.Node)
	if err != nil {
		return nil, err
	}
	c.Error = c.trace(&c.App.Model.Node)
	return c, nil
}

func (c *Context) Validate() error {
	for _, path := range c.Path {
		if err := checkMissingFlags(path.Flags); err != nil {
			return err
		}
	}
	// Check the terminal node.
	path := c.Path[len(c.Path)-1]
	switch {
	case path.App != nil:
		if err := checkMissingChildren(&path.App.Node); err != nil {
			return err
		}
		if err := checkMissingPositionals(0, path.App.Positional); err != nil {
			return err
		}

	case path.Command != nil:
		if err := checkMissingChildren(path.Command); err != nil {
			return err
		}
		if err := checkMissingPositionals(0, path.Parent.Positional); err != nil {
			return err
		}

	case path.Argument != nil:
		value := path.Argument.Argument
		if value.Required && !value.Set {
			return fmt.Errorf("%s is required", path.Argument.Summary())
		}
		if err := checkMissingChildren(path.Argument); err != nil {
			return err
		}

	case path.Positional != nil:
		if err := checkMissingPositionals(path.Positional.Position+1, path.Parent.Positional); err != nil {
			return err
		}
	}
	return nil
}

// Flags returns the accumulated available flags.
func (c *Context) Flags() (flags []*Flag) {
	for _, trace := range c.Path {
		flags = append(flags, trace.Flags...)
	}
	return
}

// Command returns the full command path.
func (c *Context) Command() (command []string) {
	for _, trace := range c.Path {
		switch {
		case trace.Positional != nil:
			command = append(command, "<"+trace.Positional.Name+">")

		case trace.Argument != nil:
			command = append(command, "<"+trace.Argument.Name+">")

		case trace.Command != nil:
			command = append(command, trace.Command.Name)
		}
	}
	return
}

// FlagValue returns the set value of a flag, if it was encountered and exists.
func (c *Context) FlagValue(flag *Flag) reflect.Value {
	for _, trace := range c.Path {
		if trace.Flag == flag {
			return trace.Value
		}
	}
	return reflect.Value{}
}

// Recursively reset values to defaults (as specified in the grammar) or the zero value.
func (c *Context) reset(node *Node) error {
	c.scan = Scan(c.args...)
	for _, flag := range node.Flags {
		err := flag.Value.Reset()
		if err != nil {
			return err
		}
	}
	for _, pos := range node.Positional {
		err := pos.Reset()
		if err != nil {
			return err
		}
	}
	for _, branch := range node.Children {
		if branch.Argument != nil {
			arg := branch.Argument
			err := arg.Reset()
			if err != nil {
				return err
			}
		}
		err := c.reset(branch)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) trace(node *Node) (err error) { // nolint: gocyclo
	positional := 0

	flags := append(c.Flags(), node.Flags...)

	for !c.scan.Peek().IsEOL() {
		token := c.scan.Peek()
		switch token.Type {
		case UntypedToken:
			switch {
			// Indicates end of parsing. All remaining arguments are treated as positional arguments only.
			case token.Value == "--":
				c.scan.Pop()
				args := []string{}
				for {
					token = c.scan.Pop()
					if token.Type == EOLToken {
						break
					}
					args = append(args, token.Value)
				}
				// Note: tokens must be pushed in reverse order.
				for i := range args {
					c.scan.PushTyped(args[len(args)-1-i], PositionalArgumentToken)
				}

				// Long flag.
			case strings.HasPrefix(token.Value, "--"):
				c.scan.Pop()
				// Parse it and push the tokens.
				parts := strings.SplitN(token.Value[2:], "=", 2)
				if len(parts) > 1 {
					c.scan.PushTyped(parts[1], FlagValueToken)
				}
				c.scan.PushTyped(parts[0], FlagToken)

				// Short flag.
			case strings.HasPrefix(token.Value, "-"):
				c.scan.Pop()
				// Note: tokens must be pushed in reverse order.
				if tail := token.Value[2:]; tail != "" {
					c.scan.PushTyped(tail, ShortFlagTailToken)
				}
				c.scan.PushTyped(token.Value[1:2], ShortFlagToken)

			default:
				c.scan.Pop()
				c.scan.PushTyped(token.Value, PositionalArgumentToken)
			}

		case ShortFlagTailToken:
			c.scan.Pop()
			// Note: tokens must be pushed in reverse order.
			if tail := token.Value[1:]; tail != "" {
				c.scan.PushTyped(tail, ShortFlagTailToken)
			}
			c.scan.PushTyped(token.Value[0:1], ShortFlagToken)

		case FlagToken:
			if err := c.matchFlags(flags, func(f *Flag) bool { return f.Name == token.Value }); err != nil {
				return err
			}

		case ShortFlagToken:
			if err := c.matchFlags(flags, func(f *Flag) bool { return string(f.Short) == token.Value }); err != nil {
				return err
			}

		case FlagValueToken:
			return fmt.Errorf("unexpected flag argument %q", token.Value)

		case PositionalArgumentToken:
			// Ensure we've consumed all positional arguments.
			if positional < len(node.Positional) {
				arg := node.Positional[positional]
				value, err := arg.Parse(c.scan)
				if err != nil {
					return err
				}
				c.Path = append(c.Path, &Path{
					Parent:     node,
					Positional: arg,
					Value:      value,
					Flags:      node.Flags,
				})
				positional++
				break
			}

			// After positional arguments have been consumed, check commands next...
			for _, branch := range node.Children {
				if branch.Type == CommandNode && branch.Name == token.Value {
					c.scan.Pop()
					c.Path = append(c.Path, &Path{
						Parent:  node,
						Command: branch,
						Value:   branch.Target,
						Flags:   node.Flags,
					})
					return c.trace(branch)
				}
			}

			// Finally, check arguments.
			for _, branch := range node.Children {
				if branch.Type == ArgumentNode {
					arg := branch.Argument
					if value, err := arg.Parse(c.scan); err == nil {
						c.Path = append(c.Path, &Path{
							Parent:   node,
							Argument: branch,
							Value:    value,
							Flags:    node.Flags,
						})
						return c.trace(branch)
					}
				}
			}
			return fmt.Errorf("unexpected positional argument %s", token)

		default:
			return fmt.Errorf("unexpected token %s", token)
		}
	}
	return nil
}

// Apply traced context to the target grammar.
func (c *Context) Apply() (string, error) {
	path := []string{}
	possibleFlags := []*Flag{}

	for _, trace := range c.Path {
		switch {
		case trace.App != nil:
			possibleFlags = append(possibleFlags, trace.App.Flags...)
		case trace.Argument != nil:
			path = append(path, "<"+trace.Argument.Name+">")
			trace.Argument.Argument.Apply(trace.Value)
			possibleFlags = append(possibleFlags, trace.Argument.Flags...)
		case trace.Command != nil:
			path = append(path, trace.Command.Name)
			possibleFlags = append(possibleFlags, trace.Command.Flags...)
		case trace.Flag != nil:
			trace.Flag.Value.Apply(trace.Value)
		case trace.Positional != nil:
			path = append(path, "<"+trace.Positional.Name+">")
			trace.Positional.Apply(trace.Value)
		}
	}

	c.applyResolvers(possibleFlags)

	return strings.Join(path, " "), nil
}

func (c *Context) applyResolvers(possibleFlags []*Flag) error {
	for _, resolver := range c.App.resolvers {
		for _, flag := range possibleFlags {
			if flag.Set {
				continue
			}

			s, err := resolver(flag)
			if err != nil {
				return err
			}

			// Create a scanner for decoding the resolved value.
			ctx := DecoderContext{Value: flag.Value}
			scan := Scanner{args: []Token{{Type: FlagValueToken, Value: s}}}

			flag.Mapper.Decode(&ctx, &scan, flag.Value.Value)
		}
	}

	return nil
}

func (c *Context) matchFlags(flags []*Flag, matcher func(f *Flag) bool) (err error) {
	defer catch(&err)
	token := c.scan.Peek()
	for _, flag := range flags {
		// Found a matching flag.
		if matcher(flag) {
			c.scan.Pop()
			value, err := flag.Parse(c.scan)
			if err != nil {
				return err
			}
			c.Path = append(c.Path, &Path{Flag: flag, Value: value})
			return nil
		}
	}
	return fmt.Errorf("unknown flag --%s", token.Value)
}

func checkMissingFlags(flags []*Flag) error {
	missing := []string{}
	for _, flag := range flags {
		if !flag.Required || flag.Set {
			continue
		}
		missing = append(missing, flag.Summary())
	}
	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf("missing flags: %s", strings.Join(missing, ", "))
}

func checkMissingChildren(node *Node) error {
	missing := []string{}
	for _, arg := range node.Positional {
		if arg.Required && !arg.Set {
			missing = append(missing, strconv.Quote(arg.Summary()))
		}
	}
	for _, child := range node.Children {
		if child.Argument != nil {
			if !child.Argument.Required {
				continue
			}
			missing = append(missing, strconv.Quote(child.Summary()))
		} else {
			missing = append(missing, strconv.Quote(child.Name))
		}
	}
	if len(missing) == 0 {
		return nil
	}

	if len(missing) == 1 {
		return fmt.Errorf("%q should be followed by %s", node.Path(), missing[0])
	}
	return fmt.Errorf("%q should be followed by one of %s", node.Path(), strings.Join(missing, ", "))
}

// If we're missing any positionals and they're required, return an error.
func checkMissingPositionals(positional int, values []*Value) error {
	// All the positionals are in.
	if positional >= len(values) {
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
