package kong

import (
	"reflect"
	"strconv"
	"strings"
)

type Application struct {
	Node
	HelpFlag *Flag
}

// Leaves returns the leaf commands/arguments in the command-line grammar.
func (a *Application) Leaves() (out []*Node) {
	var walk func(n *Node)
	walk = func(n *Node) {
		if len(n.Children) == 0 {
			out = append(out, n)
		}
		for _, child := range n.Children {
			if child.Type == CommandNode || child.Type == ArgumentNode {
				walk(child)
			}
		}
	}
	walk(&a.Node)
	return
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
	Flag     bool // True if flag, false if positional argument.
	Name     string
	Help     string
	Default  string
	Decoder  Decoder
	Tag      *Tag
	Value    reflect.Value
	Required bool
	Set      bool   // Used with Required to test if a value has been given.
	Format   string // Formatting directive, if applicable.
	Position int    // Position (for positional arguments).
}

func (v *Value) IsBool() bool {
	return v.Value.Kind() == reflect.Bool
}

// Parse tokens into value, parse, and validate, but do not write to the field.
func (v *Value) Parse(scan *Scanner) (reflect.Value, error) {
	value := reflect.New(v.Value.Type()).Elem()
	err := v.Decoder.Decode(&DecoderContext{Value: v}, scan, value)
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
	Value
	PlaceHolder string
	Env         string
	Short       rune
	Hidden      bool
}

func (f *Flag) FormatPlaceHolder() string {
	if f.PlaceHolder != "" {
		return f.PlaceHolder
	}
	if f.Default != "" {
		ellipsis := ""
		if len(f.Default) > 1 {
			ellipsis = "..."
		}

		if f.Value.Value.Kind() == reflect.String {
			return strconv.Quote(f.Default) + ellipsis
		}
		return f.Default + ellipsis
	}
	return strings.ToUpper(f.Name)
}
