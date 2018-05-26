package kong

import "reflect"

type Application struct {
	Node
	HelpFlag *Flag
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
	Target     reflect.Value
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
