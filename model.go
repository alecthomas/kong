package kong

import "reflect"

type Application = Node

// A Branch is a command or positional argument that results in a branch in the command tree.
type Branch struct {
	Command  *Command
	Argument *Argument
}

type Command = Node

type Node struct {
	Name       string
	Help       string
	Flags      []*Flag
	Positional []*Value
	Children   []*Branch
}

type Value struct {
	Name     string
	Help     string
	Decoder  Decoder
	Field    reflect.StructField
	Value    reflect.Value
	Required bool
	Format   string // Formatting directive, if applicable.
}

type Positional = Value

type Argument struct {
	Node
	Argument *Value
}

type Flag struct {
	Value
	Placeholder string
	Env         string
	Short       rune
	Default     string
}
