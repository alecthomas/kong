package kong

import (
	"fmt"
	"reflect"
	"strings"
)

func build(ast interface{}) (app *Application, err error) {
	defer func() {
		msg := recover()
		if test, ok := msg.(error); ok {
			app = nil
			err = test
		} else if msg != nil {
			panic(msg)
		}
	}()
	v := reflect.ValueOf(ast)
	iv := reflect.Indirect(v)
	if v.Kind() != reflect.Ptr || iv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a pointer to a struct but got %T", ast)
	}

	app = &Application{
		// Synthesize a --help flag.
		HelpFlag: &Flag{
			Value: Value{
				Name:    "help",
				Help:    "Show context-sensitive help.",
				Flag:    true,
				Value:   reflect.New(reflect.TypeOf(false)).Elem(),
				Decoder: kindDecoders[reflect.Bool],
			}},
	}
	node := buildNode(iv, map[string]bool{"help": true}, true)
	if len(node.Positional) > 0 && len(node.Children) > 0 {
		return nil, fmt.Errorf("can't mix positional arguments and branching arguments on %T", ast)
	}
	// Prepend --help flag.
	node.Flags = append([]*Flag{app.HelpFlag}, node.Flags...)
	app.Node = *node
	return app, nil
}

func dashedString(s string) string {
	return strings.Join(camelCase(s), "-")
}

type nodeBuilder struct {
	node      *Node
	v         reflect.Value
	seenFlags map[string]bool
	cmd       bool

	// These are specific to each field.
	i       int
	ft      reflect.StructField
	fv      reflect.Value
	name    string
	tag     *Tag
	decoder Decoder
}

func buildNode(v reflect.Value, seenFlags map[string]bool, cmd bool) *Node {
	nb := nodeBuilder{
		node:      &Node{},
		v:         v,
		seenFlags: seenFlags,
		cmd:       cmd,
	}
	return nb.build()
}

func (nb *nodeBuilder) isPrivate() bool {
	c := nb.ft.Name[0:1]
	return strings.ToLower(c) == c
}

func (nb *nodeBuilder) isStruct() bool {
	return nb.ft.Type.Kind() == reflect.Struct
}

func (nb *nodeBuilder) prepare(i int) bool {
	nb.ft = nb.v.Type().Field(i)
	if nb.isPrivate() {
		return false
	}

	nb.fv = nb.v.Field(i)

	nb.name = nb.ft.Tag.Get("name")
	if nb.name == "" {
		nb.name = strings.ToLower(dashedString(nb.ft.Name))
	}

	tag, err := parseTag(nb.fv, nb.ft.Tag.Get("kong"))
	if err != nil {
		fail("%s", err)
	}
	nb.tag = tag

	nb.decoder = DecoderForField(tag.Type, nb.ft)

	if !nb.cmd {
		nb.cmd = tag.Cmd
	}

	return true
}

func (nb *nodeBuilder) build() *Node {
	for i := 0; i < nb.v.NumField(); i++ {
		if !nb.prepare(i) {
			continue
		}

		// Nested structs are either commands or args.
		if nb.isStruct() && (nb.cmd || nb.tag.Arg) {
			nb.addChild()
		} else {
			nb.addField()
		}
	}

	// "Unsee" flags.
	for _, flag := range nb.node.Flags {
		delete(nb.seenFlags, flag.Name)
	}

	// Scan through argument positionals to ensure optional is never before a required.
	last := true
	for _, p := range nb.node.Positional {
		if !last && p.Required {
			fail("argument %q can not be required after an optional", p.Name)
		}

		last = p.Required
	}

	return nb.node
}

func (nb *nodeBuilder) addChild() {
	child := buildNode(nb.fv, nb.seenFlags, false)
	child.Help = nb.tag.Help

	// A branching argument. This is a bit hairy, as we let buildNode() do the parsing, then check that
	// a positional argument is provided to the child, and move it to the branching argument field.
	var branch *Branch
	if nb.tag.Arg {
		branch = nb.argBranch(child)
	} else {
		child.Name = nb.name
		branch = &Branch{Command: child}
	}
	nb.node.Children = append(nb.node.Children, branch)

	if len(child.Positional) > 0 && len(child.Children) > 0 {
		fail("can't mix positional arguments and branching arguments on %s.%s", nb.v.Type().Name(), nb.ft.Name)
	}
}

func (nb *nodeBuilder) argBranch(child *Node) *Branch {
	if len(child.Positional) == 0 {
		fail("positional branch %s.%s must have at least one child positional argument",
			nb.v.Type().Name(), nb.ft.Name)
	}

	value := child.Positional[0]
	child.Positional = child.Positional[1:]
	if child.Help == "" {
		child.Help = value.Help
	}

	child.Name = value.Name
	if child.Name != nb.name {
		fail("first field in positional branch %s.%s must have the same name as the parent field (%s).",
			nb.v.Type().Name(), nb.ft.Name, child.Name)
	}

	return &Branch{Argument: &Argument{
		Node:     *child,
		Argument: value,
	}}
}

func (nb *nodeBuilder) addField() {
	if nb.decoder == nil {
		fail("no decoder for %s.%s (of type %s)", nb.v.Type(), nb.ft.Name, nb.ft.Type)
	}

	flag := !nb.tag.Arg
	value := Value{
		Name:    nb.name,
		Flag:    flag,
		Help:    nb.tag.Help,
		Default: nb.tag.Default,
		Decoder: nb.decoder,
		Value:   nb.fv,
		Format:  nb.tag.Format,

		// Flags are optional by default, and args are required by default.
		Required: (flag && nb.tag.Required) || (nb.tag.Arg && !nb.tag.Optional),
	}

	if nb.tag.Arg {
		nb.node.Positional = append(nb.node.Positional, &value)
	} else {
		if nb.seenFlags[value.Name] {
			fail("duplicate flag --%s", value.Name)
		}
		nb.seenFlags[value.Name] = true
		nb.node.Flags = append(nb.node.Flags, &Flag{
			Value:       value,
			Short:       nb.tag.Short,
			Placeholder: nb.tag.Placeholder,
			Env:         nb.tag.Env,
		})
	}
}
