package kong

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Application struct {
	Node
	HelpFlag *Flag
}

type Argument = Node

type Command = Node

type NodeType int

const (
	ApplicationNode NodeType = iota
	CommandNode
	ArgumentNode
)

type Node struct {
	Type       NodeType
	Parent     *Node
	Name       string
	Help       string
	Flags      []*Flag
	Positional []*Positional
	Children   []*Node
	Target     reflect.Value // Pointer to the value in the grammar that this Node is associated with.

	Argument *Value // Populated when Type is ArgumentNode.
}

func (n *Node) AllFlags() (out [][]*Flag) {
	if n.Parent != nil {
		out = append(out, n.Parent.AllFlags()...)
	}
	if len(n.Flags) > 0 {
		out = append(out, n.Flags)
	}
	return
}

// Leaves returns the leaf commands/arguments under Node.
func (n *Node) Leaves() (out []*Node) {
	var walk func(n *Node)
	walk = func(n *Node) {
		if len(n.Children) == 0 && n.Type != ApplicationNode {
			out = append(out, n)
		}
		for _, child := range n.Children {
			if child.Type == CommandNode || child.Type == ArgumentNode {
				walk(child)
			}
		}
	}
	for _, child := range n.Children {
		walk(child)
	}
	return
}

// Depth of the command from the application root.
func (n *Node) Depth() int {
	depth := 0
	p := n.Parent
	for p != nil && p.Type != ApplicationNode {
		depth++
		p = p.Parent
	}
	return depth
}

// Summary help string for the node.
func (n *Node) Summary() string {
	summary := n.Path()
	if flags := n.FlagSummary(); flags != "" {
		summary += " " + flags
	}
	args := []string{}
	for _, arg := range n.Positional {
		args = append(args, arg.Summary())
	}
	if len(args) != 0 {
		summary += " " + strings.Join(args, " ")
	} else if len(n.Children) > 0 {
		summary += " <command>"
	}
	return summary
}

// FlagSummary for the node.
func (n *Node) FlagSummary() string {
	required := []string{}
	count := 0
	for _, group := range n.AllFlags() {
		for _, flag := range group {
			count++
			if flag.Required {
				required = append(required, flag.Summary())
			}
		}
	}
	return strings.Join(required, " ")
}

// Path through ancestors to this Node.
func (n *Node) Path() (out string) {
	if n.Parent != nil {
		out += " " + n.Parent.Path()
	}
	switch n.Type {
	case ApplicationNode, CommandNode:
		out += " " + n.Name
	case ArgumentNode:
		out += " " + "<" + n.Name + ">"
	}
	return strings.TrimSpace(out)
}

// A Value is either a flag or a variable positional argument.
type Value struct {
	Flag     *Flag
	Name     string
	Help     string
	Default  string
	Mapper   Mapper
	Tag      *Tag
	Value    reflect.Value
	Required bool
	Set      bool   // Set to true when this value is set through some mechanism.
	Format   string // Formatting directive, if applicable.
	Position int    // Position (for positional arguments).
}

func (v *Value) Summary() string {
	if v.Flag != nil {
		if v.IsBool() {
			return fmt.Sprintf("--%s", v.Name)
		}
		return fmt.Sprintf("--%s=%s", v.Name, v.Flag.FormatPlaceHolder())
	}
	argText := "<" + v.Name + ">"
	if v.IsCumulative() {
		argText += " ..."
	}
	if !v.Required {
		argText = "[" + argText + "]"
	}
	return argText
}

func (v *Value) IsCumulative() bool {
	return v.Value.Kind() == reflect.Slice
}

func (v *Value) IsBool() bool {
	if m, ok := v.Mapper.(BoolMapper); ok && m.IsBool() {
		return true
	}
	return v.Value.Kind() == reflect.Bool
}

// Parse tokens into value, parse, and validate, but do not write to the field.
func (v *Value) Parse(scan *Scanner) (reflect.Value, error) {
	value := reflect.New(v.Value.Type()).Elem()
	err := v.Mapper.Decode(&DecoderContext{Value: v}, scan, value)
	if err == nil {
		v.Set = true
	}
	return value, err
}

// Apply value to field.
func (v *Value) Apply(value reflect.Value) {
	v.Value.Set(value)
	v.Set = true
}

func (v *Value) Reset() error {
	v.Value.Set(reflect.Zero(v.Value.Type()))
	if v.Default != "" {
		value, err := v.Parse(Scan(v.Default))
		if err != nil {
			return err
		}
		v.Apply(value)
		v.Set = false
	}
	return nil
}

type Positional = Value

type Flag struct {
	*Value
	PlaceHolder string
	Env         string
	Short       rune
	Hidden      bool
}

func (f *Flag) String() string {
	out := "--" + f.Name
	if f.Short != 0 {
		out = fmt.Sprintf("-%c, %s", f.Short, out)
	}
	if !f.IsBool() {
		out += "=" + f.FormatPlaceHolder()
	}
	return out
}

func (f *Flag) FormatPlaceHolder() string {
	tail := ""
	if f.Value.IsCumulative() {
		tail += ", ..."
	}
	if f.Default != "" {
		if f.Value.Value.Kind() == reflect.String {
			return strconv.Quote(f.Default) + tail
		}
		return f.Default + tail
	}
	if f.PlaceHolder != "" {
		return f.PlaceHolder + tail
	}
	return strings.ToUpper(f.Name) + tail
}
