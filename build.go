package kong

import (
	"fmt"
	"reflect"
	"strings"
)

func build(ast interface{}, extraFlags []*Flag) (app *Application, err error) {
	defer catch(&err)
	v := reflect.ValueOf(ast)
	iv := reflect.Indirect(v)
	if v.Kind() != reflect.Ptr || iv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a pointer to a struct but got %T", ast)
	}

	app = &Application{}
	seenFlags := map[string]bool{}
	for _, flag := range extraFlags {
		seenFlags[flag.Name] = true
	}
	node := buildNode(iv, ApplicationNode, seenFlags)
	if len(node.Positional) > 0 && len(node.Children) > 0 {
		return nil, fmt.Errorf("can't mix positional arguments and branching arguments on %T", ast)
	}
	app.Node = *node
	app.Node.Flags = append(extraFlags, app.Node.Flags...)
	return app, nil
}

func dashedString(s string) string {
	return strings.Join(camelCase(s), "-")
}

func buildNode(v reflect.Value, typ NodeType, seenFlags map[string]bool) *Node {
	node := &Node{
		Type:   typ,
		Target: v,
	}
	for i := 0; i < v.NumField(); i++ {
		ft := v.Type().Field(i)
		if strings.ToLower(ft.Name[0:1]) == ft.Name[0:1] {
			continue
		}
		fv := v.Field(i)

		name := ft.Tag.Get("name")
		if name == "" {
			name = strings.ToLower(dashedString(ft.Name))
		}

		tag := parseTag(fv, ft.Tag.Get("kong"))

		// Nested structs are either commands or args.
		if ft.Type.Kind() == reflect.Struct && (tag.Cmd || tag.Arg) {
			typ := CommandNode
			if tag.Arg {
				typ = ArgumentNode
			}
			buildChild(node, typ, v, ft, fv, tag, name, seenFlags)
		} else {
			buildField(node, v, ft, fv, tag, name, seenFlags)
		}
	}

	// "Unsee" flags.
	for _, flag := range node.Flags {
		delete(seenFlags, flag.Name)
	}

	// Scan through argument positionals to ensure optional is never before a required.
	last := true
	for i, p := range node.Positional {
		if !last && p.Required {
			fail("argument %q can not be required after an optional", p.Name)
		}

		last = p.Required
		p.Position = i
	}

	return node
}

func buildChild(node *Node, typ NodeType, v reflect.Value, ft reflect.StructField, fv reflect.Value, tag *Tag, name string, seenFlags map[string]bool) {
	child := buildNode(fv, typ, seenFlags)
	child.Parent = node
	child.Help = tag.Help

	// A branching argument. This is a bit hairy, as we let buildNode() do the parsing, then check that
	// a positional argument is provided to the child, and move it to the branching argument field.
	if tag.Arg {
		if len(child.Positional) == 0 {
			fail("positional branch %s.%s must have at least one child positional argument",
				v.Type().Name(), ft.Name)
		}

		value := child.Positional[0]
		child.Positional = child.Positional[1:]
		if child.Help == "" {
			child.Help = value.Help
		}

		child.Name = value.Name
		if child.Name != name {
			fail("first field in positional branch %s.%s must have the same name as the parent field (%s).",
				v.Type().Name(), ft.Name, child.Name)
		}

		child.Argument = value
	} else {
		child.Name = name
	}
	node.Children = append(node.Children, child)

	if len(child.Positional) > 0 && len(child.Children) > 0 {
		fail("can't mix positional arguments and branching arguments on %s.%s", v.Type().Name(), ft.Name)
	}
}

func buildField(node *Node, v reflect.Value, ft reflect.StructField, fv reflect.Value, tag *Tag, name string, seenFlags map[string]bool) {
	decoder := DecoderForField(tag.Type, ft)
	if decoder == nil {
		fail("no decoder for %s.%s (of type %s)", v.Type(), ft.Name, ft.Type)
	}

	flag := !tag.Arg

	value := Value{
		Name:    name,
		Flag:    flag,
		Help:    tag.Help,
		Default: tag.Default,
		Decoder: decoder,
		Tag:     tag,
		Value:   fv,

		// Flags are optional by default, and args are required by default.
		Required: (flag && tag.Required) || (tag.Arg && !tag.Optional),
		Format:   tag.Format,
	}

	if tag.Arg {
		node.Positional = append(node.Positional, &value)
	} else {
		if seenFlags[value.Name] {
			fail("duplicate flag --%s", value.Name)
		}
		seenFlags[value.Name] = true
		node.Flags = append(node.Flags, &Flag{
			Value:       value,
			Short:       tag.Short,
			PlaceHolder: tag.PlaceHolder,
			Env:         tag.Env,
		})
	}
}
