package kong

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Application is the root of the Kong model.
type Application struct {
	*Node
	// Help flag, if the NoDefaultHelp() option is not specified.
	HelpFlag *Flag
}

// Argument represents a branching positional argument.
type Argument = Node

// Command represents a command in the CLI.
type Command = Node

// NodeType is an enum representing the type of a Node.
type NodeType int

// Node type enumerations.
const (
	ApplicationNode NodeType = iota
	CommandNode
	ArgumentNode
)

// Node is a branch in the CLI. ie. a command or positional argument.
type Node struct {
	Type       NodeType
	Parent     *Node
	Name       string
	Help       string
	Hidden     bool
	Flags      []*Flag
	Positional []*Positional
	Children   []*Node
	Target     reflect.Value // Pointer to the value in the grammar that this Node is associated with.

	Argument *Value // Populated when Type is ArgumentNode.
}

// Leaf returns true if this Node is a leaf node.
func (n *Node) Leaf() bool {
	return len(n.Children) == 0
}

// Find a command/argument/flag by pointer to its field.
//
// Returns nil if not found. Panics if ptr is not a pointer.
func (n *Node) Find(ptr interface{}) *Node {
	key := reflect.ValueOf(ptr)
	if key.Kind() != reflect.Ptr {
		panic("expected a pointer")
	}
	return n.findNode(key)
}

func (n *Node) findNode(key reflect.Value) *Node {
	if n.Target == key {
		return n
	}
	for _, child := range n.Children {
		if found := child.findNode(key); found != nil {
			return found
		}
	}
	return nil
}

// AllFlags returns flags from all ancestor branches encountered.
//
// If "hide" is true hidden flags will be omitted.
func (n *Node) AllFlags(hide bool) (out [][]*Flag) {
	if n.Parent != nil {
		out = append(out, n.Parent.AllFlags(hide)...)
	}
	group := []*Flag{}
	for _, flag := range n.Flags {
		if !hide || !flag.Hidden {
			group = append(group, flag)
		}
	}
	if len(group) > 0 {
		out = append(out, group)
	}
	return
}

// Leaves returns the leaf commands/arguments under Node.
//
// If "hidden" is true hidden leaves will be omitted.
func (n *Node) Leaves(hide bool) (out []*Node) {
	var walk func(n *Node)
	walk = func(n *Node) {
		if hide && n.Hidden {
			return
		}
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

// Summary help string for the node (not including application name).
func (n *Node) Summary() string {
	summary := n.Path()
	if flags := n.FlagSummary(true); flags != "" {
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
func (n *Node) FlagSummary(hide bool) string {
	required := []string{}
	count := 0
	for _, group := range n.AllFlags(hide) {
		for _, flag := range group {
			count++
			if flag.Required {
				required = append(required, flag.Summary())
			}
		}
	}
	return strings.Join(required, " ")
}

// FullPath is like Path() but includes the Application root node.
func (n *Node) FullPath() string {
	root := n
	for root.Parent != nil {
		root = root.Parent
	}
	return strings.TrimSpace(root.Name + " " + n.Path())
}

// Path through ancestors to this Node.
func (n *Node) Path() (out string) {
	if n.Parent != nil {
		out += " " + n.Parent.Path()
	}
	switch n.Type {
	case CommandNode:
		out += " " + n.Name
	case ArgumentNode:
		out += " " + "<" + n.Name + ">"
	}
	return strings.TrimSpace(out)
}

// A Value is either a flag or a variable positional argument.
type Value struct {
	Flag     *Flag // Nil if positional argument.
	Name     string
	Help     string
	Default  string
	Enum     string
	Mapper   Mapper
	Tag      *Tag
	Target   reflect.Value
	Required bool
	Set      bool   // Set to true when this value is set through some mechanism.
	Format   string // Formatting directive, if applicable.
	Position int    // Position (for positional arguments).
}

// EnumMap returns a map of the enums in this value.
func (v *Value) EnumMap() map[string]bool {
	parts := strings.Split(v.Enum, ",")
	out := make(map[string]bool, len(parts))
	for _, part := range parts {
		out[strings.TrimSpace(part)] = true
	}
	return out
}

// Summary returns a human-readable summary of the value.
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

// IsCumulative returns true if the type can be accumulated into.
func (v *Value) IsCumulative() bool {
	return v.IsSlice() || v.IsMap()
}

// IsSlice returns true if the value is a slice.
func (v *Value) IsSlice() bool {
	return v.Target.Kind() == reflect.Slice
}

// IsMap returns true if the value is a map.
func (v *Value) IsMap() bool {
	return v.Target.Kind() == reflect.Map
}

// IsBool returns true if the underlying value is a boolean.
func (v *Value) IsBool() bool {
	if m, ok := v.Mapper.(BoolMapper); ok && m.IsBool() {
		return true
	}
	return v.Target.Kind() == reflect.Bool
}

// Parse tokens into value, parse, and validate, but do not write to the field.
func (v *Value) Parse(scan *Scanner, target reflect.Value) error {
	err := v.Mapper.Decode(&DecodeContext{Value: v, Scan: scan}, target)
	if err == nil {
		v.Set = true
	}
	return err
}

// Apply value to field.
func (v *Value) Apply(value reflect.Value) {
	v.Target.Set(value)
	v.Set = true
}

// Reset this value to its default, either the zero value or the parsed result of its "default" tag.
//
// Does not include resolvers.
func (v *Value) Reset() error {
	v.Target.Set(reflect.Zero(v.Target.Type()))
	if v.Default != "" {
		return v.Parse(Scan(v.Default), v.Target)
	}
	return nil
}

// A Positional represents a non-branching command-line positional argument.
type Positional = Value

// A Flag represents a command-line flag.
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

// FormatPlaceHolder formats the placeholder string for a Flag.
func (f *Flag) FormatPlaceHolder() string {
	tail := ""
	if f.Value.IsSlice() {
		tail += ",..."
	}
	if f.Default != "" {
		if f.Value.Target.Kind() == reflect.String {
			return strconv.Quote(f.Default) + tail
		}
		return f.Default + tail
	}
	if f.PlaceHolder != "" {
		return f.PlaceHolder + tail
	}
	if f.Value.IsMap() {
		return "KEY=VALUE" + tail
	}
	return strings.ToUpper(f.Name) + tail
}
