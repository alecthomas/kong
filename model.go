package kong

import "reflect"

type Application struct {
	Node
	HelpFlag bool // Target of --help flag.
}

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

// A Value is either a flag or a variable positional argument.
type Value struct {
	Flag     bool // True if flag, false if positional argument.
	Name     string
	Help     string
	Default  string
	Decoder  Decoder
	Field    reflect.StructField
	Value    reflect.Value
	Required bool
	Set      bool   // Used with Required to test if a value has been given.
	Format   string // Formatting directive, if applicable.
}

func (v *Value) Decode(scan *Scanner) error {
	err := v.Decoder.Decode(&DecoderContext{Value: v}, scan, v.Value)
	if err == nil {
		v.Set = true
	}
	return err
}

func (v *Value) Reset() {
	v.Value.Set(reflect.Zero(v.Value.Type()))
	if v.Default != "" {
		v.Decode(Scan(v.Default))
		v.Set = false
	}
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
}
