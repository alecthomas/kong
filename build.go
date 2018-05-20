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

func buildNode(v reflect.Value, seenFlags map[string]bool, cmd bool) *Node {
	node := &Node{}
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

		tag, err := parseTag(fv, ft.Tag.Get("kong"))
		if err != nil {
			fail("%s", err)
		}

		decoder := DecoderForField(tag.Type, ft)

		if !cmd {
			cmd = tag.Cmd
		}

		env := ft.Tag.Get("env")
		format := ft.Tag.Get("format")

		// Nested structs are either commands or args.
		if ft.Type.Kind() == reflect.Struct && (cmd || tag.Arg) {
			child := buildNode(fv, seenFlags, false)
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
				node.Children = append(node.Children, &Branch{Argument: &Argument{
					Node:     *child,
					Argument: value,
				}})
			} else {
				child.Name = name
				node.Children = append(node.Children, &Branch{Command: child})
			}

			if len(child.Positional) > 0 && len(child.Children) > 0 {
				fail("can't mix positional arguments and branching arguments on %s.%s", v.Type().Name(), ft.Name)
			}
		} else {
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
				Value:   fv,

				// Flags are optional by default, and args are required by default.
				Required: (flag && tag.Required) || (tag.Arg && !tag.Optional),
				Format:   format,
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
					Placeholder: tag.Placeholder,
					Env:         env,
				})
			}
		}
	}

	// "Unsee" flags.
	for _, flag := range node.Flags {
		delete(seenFlags, flag.Name)
	}

	// Scan through argument positionals to ensure optional is never before a required.
	last := true
	for _, p := range node.Positional {
		if !last && p.Required {
			fail("argument %q can not be required after an optional", p.Name)
		}

		last = p.Required
	}

	return node
}
