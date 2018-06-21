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

	// True if this Path element was created as the result of a resolver.
	Resolved bool
}

// Node returns the Node associated with this Path, or nil if Path is a non-Node.
func (p *Path) Node() *Node {
	switch {
	case p.App != nil:
		return &p.App.Node

	case p.Argument != nil:
		return p.Argument

	case p.Command != nil:
		return p.Command
	}
	return nil
}

// Context contains the current parse context.
type Context struct {
	App   *Kong
	Path  []*Path  // A trace through parsed nodes.
	Args  []string // Original command-line arguments.
	Error error    // Error that occurred during trace, if any.

	values map[*Value]reflect.Value // Temporary values during tracing.
	scan   *Scanner
}

// Value returns the value for a particular path element.
func (c *Context) Value(path *Path) reflect.Value {
	switch {
	case path.Positional != nil:
		return c.values[path.Positional]
	case path.Flag != nil:
		return c.values[path.Flag.Value]
	case path.Argument != nil:
		return c.values[path.Argument.Argument]
	}
	panic("can only retrieve value for flag, argument or positional")
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
//
// Note that this will not modify the target grammar. Call Apply() to do so.
func Trace(k *Kong, args []string) (*Context, error) {
	c := &Context{
		App:  k,
		Args: args,
		Path: []*Path{
			{App: k.Model, Flags: k.Model.Flags},
		},
		values: map[*Value]reflect.Value{},
		scan:   Scan(args...),
	}
	c.Error = c.trace(&c.App.Model.Node)
	return c, c.traceResolvers()
}

// Validate the current context.
func (c *Context) Validate() error {
	for _, path := range c.Path {
		if err := checkMissingFlags(path.Flags); err != nil {
			return err
		}
	}
	// Check the terminal node.
	node := c.Selected()
	if node == nil {
		node = &c.App.Model.Node
	}

	// Find deepest positional argument so we can check if all required positionals have been provided.
	positionals := 0
	for _, path := range c.Path {
		if path.Positional != nil {
			positionals = path.Positional.Position + 1
		}
	}

	if err := checkMissingChildren(node); err != nil {
		return err
	}
	if err := checkMissingPositionals(positionals, node.Positional); err != nil {
		return err
	}

	if node.Type == ArgumentNode {
		value := node.Argument
		if value.Required && !value.Set {
			return fmt.Errorf("%s is required", node.Summary())
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
			return c.values[trace.Flag.Value]
		}
	}
	return reflect.Value{}
}

// Recursively reset values to defaults (as specified in the grammar) or the zero value.
func (c *Context) reset(node *Node) error {
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
			if err := c.parseFlag(flags, "--"+token.Value); err != nil {
				return err
			}

		case ShortFlagToken:
			if err := c.parseFlag(flags, "-"+token.Value); err != nil {
				return err
			}

		case FlagValueToken:
			return fmt.Errorf("unexpected flag argument %q", token.Value)

		case PositionalArgumentToken:
			// Ensure we've consumed all positional arguments.
			if positional < len(node.Positional) {
				arg := node.Positional[positional]
				err := arg.Parse(c.scan, c.getValue(arg))
				if err != nil {
					return err
				}
				c.Path = append(c.Path, &Path{
					Parent:     node,
					Positional: arg,
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
						Flags:   branch.Flags,
					})
					return c.trace(branch)
				}
			}

			// Finally, check arguments.
			for _, branch := range node.Children {
				if branch.Type == ArgumentNode {
					arg := branch.Argument
					if err := arg.Parse(c.scan, c.getValue(arg)); err == nil {
						c.Path = append(c.Path, &Path{
							Parent:   node,
							Argument: branch,
							Flags:    branch.Flags,
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

// Walk through flags from existing nodes in the path.
func (c *Context) traceResolvers() error {
	if len(c.App.resolvers) == 0 {
		return nil
	}

	inserted := []*Path{}
	for _, path := range c.Path {
		for _, flag := range path.Flags {
			// Flag has already been set on the command-line.
			if _, ok := c.values[flag.Value]; ok {
				continue
			}
			for _, resolver := range c.App.resolvers {
				s, err := resolver(c, path, flag)
				if err != nil {
					return err
				}
				if s == "" {
					continue
				}

				scan := Scan().PushTyped(s, FlagValueToken)
				delete(c.values, flag.Value)
				err = flag.Parse(scan, c.getValue(flag.Value))
				if err != nil {
					return err
				}
				inserted = append(inserted, &Path{
					Flag:     flag,
					Resolved: true,
				})
			}
		}
	}
	c.Path = append(inserted, c.Path...)
	return nil
}

func (c *Context) getValue(value *Value) reflect.Value {
	v, ok := c.values[value]
	if !ok {
		v = reflect.New(value.Target.Type()).Elem()
		c.values[value] = v
	}
	return v
}

// Apply traced context to the target grammar.
func (c *Context) Apply() (string, error) {
	err := c.reset(&c.App.Model.Node)
	if err != nil {
		return "", err
	}

	path := []string{}

	for _, trace := range c.Path {
		var value *Value
		switch {
		case trace.App != nil:
		case trace.Argument != nil:
			path = append(path, "<"+trace.Argument.Name+">")
			value = trace.Argument.Argument
		case trace.Command != nil:
			path = append(path, trace.Command.Name)
		case trace.Flag != nil:
			value = trace.Flag.Value
		case trace.Positional != nil:
			path = append(path, "<"+trace.Positional.Name+">")
			value = trace.Positional
		default:
			panic("unsupported path ?!")
		}
		if value != nil {
			value.Apply(c.getValue(value))
		}
	}

	return strings.Join(path, " "), nil
}

func (c *Context) parseFlag(flags []*Flag, match string) (err error) {
	defer catch(&err)
	for _, flag := range flags {
		if "-"+string(flag.Short) != match && "--"+flag.Name != match {
			continue
		}
		// Found a matching flag.
		c.scan.Pop()
		err := flag.Parse(c.scan, c.getValue(flag.Value))
		if err != nil {
			return err
		}
		c.Path = append(c.Path, &Path{Flag: flag})
		return nil
	}
	return fmt.Errorf("unknown flag %s", match)
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
	if len(missing) > 5 {
		missing = append(missing[:5], "...")
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
