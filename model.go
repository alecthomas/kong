package kong

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// A Visitable component in the model.
type Visitable interface {
	node()
}

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
	Help       string // Short help displayed in summaries.
	Detail     string // Detailed help displayed when describing command/arg alone.
	Group      string
	Hidden     bool
	Flags      []*Flag
	Positional []*Positional
	Children   []*Node
	Target     reflect.Value // Pointer to the value in the grammar that this Node is associated with.
	Tag        *Tag

	Argument *Value // Populated when Type is ArgumentNode.
}

func (*Node) node() {}

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
	_ = Visit(n, func(nd Visitable, next Next) error {
		if nd == n {
			return next(nil)
		}
		if node, ok := nd.(*Node); ok {
			if hide && node.Hidden {
				return nil
			}
			if len(node.Children) == 0 && node.Type != ApplicationNode {
				out = append(out, node)
			}
		}
		return next(nil)
	})
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

// Vars returns the combined Vars defined by all ancestors of this Node.
func (n *Node) Vars() Vars {
	if n == nil {
		return Vars{}
	}
	return n.Parent.Vars().CloneWith(n.Tag.Vars)
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

type completionOptions map[string]struct{}

func (c completionOptions) add(option ...string) {
	for _, opt := range option {
		c[opt] = struct{}{}
	}
}

func (c completionOptions) slice() []string {
	options := make([]string, 0, len(c))
	for s := range c {
		options = append(options, s)
	}
	sort.Strings(options)
	return options
}

func (n *Node) subCommandCompletion(ctx *Context, a CompleterArgs, opts *completionOptions) (found, stop bool, err error) {
	var subCommand *Node
	completedArgs := a.Completed()
	for i := 0; i < len(completedArgs); i++ {
		i := i
		for _, sub := range n.Children {
			if sub.Name == completedArgs[i] && sub.Type == CommandNode {
				subCommand = sub
				found = true
				break
			}
		}
		if !found {
			continue
		}
		// when there's a subCommand, run the subCommand's completion with the remaining args
		if i < len(a) {
			a = a[i+1:]
		} else {
			a = a[len(a):]
		}
		stop, err = subCommand.runCompletion(ctx, a, opts)
		if err != nil {
			return false, false, err
		}
		return found, stop, nil
	}
	return found, stop, nil
}

func (n *Node) runCompletion(ctx *Context, a CompleterArgs, opts *completionOptions) (bool, error) {
	foundSub, _, err := n.subCommandCompletion(ctx, a, opts)
	if err != nil {
		return false, err
	}

	if foundSub {
		return false, nil
	}

	stop, err := n.flagCompletion(ctx, a, opts)
	if err != nil || stop {
		return stop, err
	}

	err = n.positionalCompletion(ctx, a, opts)
	if err != nil {
		return false, err
	}

	for _, child := range n.Children {
		if child.Type == CommandNode {
			opts.add(child.Name)
		}
	}

	return false, nil
}

func (n *Node) positionalCompletion(ctx *Context, a CompleterArgs, opts *completionOptions) error {
	comp := &positionalCompleter{
		Flags:      n.Flags,
		Completers: make([]Completer, len(n.Positional)),
	}
	var err error
	for i, pos := range n.Positional {
		comp.Completers[i], err = pos.completer(ctx)
		if err != nil {
			return err
		}
	}
	opts.add(comp.Options(a)...)
	return nil
}

func (n *Node) flagCompletion(ctx *Context, a CompleterArgs, opts *completionOptions) (bool, error) {
	if n.Flags == nil {
		return false, nil
	}
	completers := make(map[string]Completer, len(n.Flags)*2)
	for _, flags := range n.AllFlags(true) {
		for _, flag := range flags {
			if flag == nil {
				continue
			}
			completer, err := flag.completer(ctx)
			if err != nil {
				return false, err
			}
			completers["--"+flag.Name] = completer
			if flag.Short == 0 {
				continue
			}
			completers["-"+string(flag.Short)] = completer
		}
	}

	completer, ok := completers[a.LastCompleted()]
	if ok && completer != nil {
		opts.add(completer.Options(a)...)
		// stop further completions when a flag is being completed
		return true, nil
	}

	if strings.HasPrefix(a.Last(), "-") {
		for flag := range completers {
			opts.add(flag)
		}
	}
	return false, nil
}

// A Value is either a flag or a variable positional argument.
type Value struct {
	Flag         *Flag // Nil if positional argument.
	Name         string
	Help         string
	Default      string
	DefaultValue reflect.Value
	Enum         string
	Mapper       Mapper
	Tag          *Tag
	Target       reflect.Value
	Required     bool
	Set          bool   // Set to true when this value is set through some mechanism.
	Format       string // Formatting directive, if applicable.
	Position     int    // Position (for positional arguments).
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

// ShortSummary returns a human-readable summary of the value, not including any placeholders/defaults.
func (v *Value) ShortSummary() string {
	if v.Flag != nil {
		return fmt.Sprintf("--%s", v.Name)
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
	return v.Target.Type().Name() == "" && v.Target.Kind() == reflect.Slice
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
func (v *Value) Parse(scan *Scanner, target reflect.Value) (err error) {
	defer func() {
		if rerr := recover(); rerr != nil {
			switch rerr := rerr.(type) {
			case Error:
				err = errors.Wrap(rerr, v.ShortSummary())
			default:
				panic(fmt.Sprintf("mapper %T failed to apply to %s: %s", v.Mapper, v.Summary(), rerr))
			}
		}
	}()
	err = v.Mapper.Decode(&DecodeContext{Value: v, Scan: scan}, target)
	if err != nil {
		return errors.Wrap(err, v.ShortSummary())
	}
	v.Set = true
	return nil
}

// Apply value to field.
func (v *Value) Apply(value reflect.Value) {
	v.Target.Set(value)
	v.Set = true
}

// ApplyDefault value to field if it is not already set.
func (v *Value) ApplyDefault() error {
	if reflectValueIsZero(v.Target) {
		return v.Reset()
	}
	v.Set = true
	return nil
}

// Reset this value to its default, either the zero value or the parsed result of its envar,
// or its "default" tag.
//
// Does not include resolvers.
func (v *Value) Reset() error {
	v.Target.Set(reflect.Zero(v.Target.Type()))
	if v.Tag.Env != "" {
		envar := os.Getenv(v.Tag.Env)
		if envar != "" {
			err := v.Parse(ScanFromTokens(Token{Type: FlagValueToken, Value: envar}), v.Target)
			if err != nil {
				return fmt.Errorf("%s (from envar %s=%q)", err, v.Tag.Env, envar)
			}
			return nil
		}
	}
	if v.Default != "" {
		return v.Parse(ScanFromTokens(Token{Type: FlagValueToken, Value: v.Default}), v.Target)
	}
	return nil
}

func (*Value) node() {}

func (v *Value) completer(ctx *Context) (Completer, error) {
	if v == nil {
		return nil, nil
	}
	completer, err := v.Tag.completer(ctx)
	if err != nil {
		return nil, err
	}
	if completer != nil {
		return completer, nil
	}
	switch {
	case v.IsBool():
		return CompleteNothing(), nil

	case v.Enum != "":
		enumVals := make([]string, 0, len(v.EnumMap()))
		for enumVal := range v.EnumMap() {
			enumVals = append(enumVals, enumVal)
		}
		return CompleteSet(enumVals...), nil

	default:
		return CompleteAnything(), nil
	}
}

// A Positional represents a non-branching command-line positional argument.
type Positional = Value

// A Flag represents a command-line flag.
type Flag struct {
	*Value
	Group       string // Logical grouping when displaying. May also be used by configuration loaders to group options logically.
	Xor         string
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
		return "KEY=VALUE;..."
	}
	return strings.ToUpper(f.Name) + tail
}

// This is directly from the Go 1.13 source code.
func reflectValueIsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return math.Float64bits(v.Float()) == 0
	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		return math.Float64bits(real(c)) == 0 && math.Float64bits(imag(c)) == 0
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !reflectValueIsZero(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		return v.IsNil()
	case reflect.String:
		return v.Len() == 0
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !reflectValueIsZero(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		// This should never happens, but will act as a safeguard for
		// later, as a default value doesn't makes sense here.
		panic(&reflect.ValueError{"reflect.Value.IsZero", v.Kind()})
	}
}
