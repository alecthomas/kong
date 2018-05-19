package kong

import (
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"
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

	node := buildNode(iv, true)
	if len(node.Positional) > 0 && len(node.Children) > 0 {
		return nil, fmt.Errorf("can't mix positional arguments and branching arguments on %T", ast)
	}
	return node, nil
}

func buildNode(v reflect.Value, cmd bool) *Node {
	node := &Node{}
	for i := 0; i < v.NumField(); i++ {
		ft := v.Type().Field(i)
		if strings.ToLower(ft.Name[0:1]) == ft.Name[0:1] {
			continue
		}
		fv := v.Field(i)

		name := ft.Tag.Get("name")
		if name == "" {
			name = strings.ToLower(strings.Join(camelCase(ft.Name), "-"))
		}
		decoder := DecoderForField(ft)
		help, _ := ft.Tag.Lookup("help")
		dflt := ft.Tag.Get("default")
		placeholder := ft.Tag.Get("placeholder")
		if placeholder == "" {
			placeholder = strings.ToUpper(strings.Join(camelCase(fv.Type().Name()), "-"))
		}
		short, _ := utf8.DecodeRuneInString(ft.Tag.Get("short"))
		if short == utf8.RuneError {
			short = 0
		}
		// group := ft.Tag.Get("group")
		_, required := ft.Tag.Lookup("required")
		_, optional := ft.Tag.Lookup("optional")
		// Force field to be an argument, not a flag.
		_, arg := ft.Tag.Lookup("arg")
		if !cmd {
			_, cmd = ft.Tag.Lookup("cmd")
		}
		env := ft.Tag.Get("env")
		format := ft.Tag.Get("format")

		// Nested structs are either commands or args.
		if ft.Type.Kind() == reflect.Struct && (cmd || arg) {
			child := buildNode(fv, false)
			child.Help = help

			// A branching argument. This is a bit hairy, as we let buildNode() do the parsing, then check that
			// a positional argument is provided to the child, and move it to the branching argument field.
			if arg {
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

			flag := !arg

			value := Value{
				Name:    name,
				Flag:    flag,
				Help:    help,
				Decoder: decoder,
				Value:   fv,
				Field:   ft,

				// Flags are optional by default, and args are required by default.
				Required: (flag && required) || (arg && !optional),
				Format:   format,
			}
			if arg {
				node.Positional = append(node.Positional, &value)
			} else {
				node.Flags = append(node.Flags, &Flag{
					Value:       value,
					Short:       short,
					Default:     dflt,
					Placeholder: placeholder,
					Env:         env,
				})
			}
		}
	}

	// Scan through argument positionals to ensure optional is never before a required
	last := true
	for _, p := range node.Positional {
		if p.Flag {
			continue
		}

		if !last && p.Required {
			fail("arguments can not be required after an optional: %v", p.Name)
		}

		last = p.Required
	}

	return node
}
