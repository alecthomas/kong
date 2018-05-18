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

	return buildNode(iv, true), nil
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
		} else {
			if decoder == nil {
				fail("no decoder for %s.%s (of type %s)", v.Type(), ft.Name, ft.Type)
			}
			value := Value{
				Name:     name,
				Help:     help,
				Decoder:  decoder,
				Value:    fv,
				Field:    ft,
				Required: !optional || required,
				Format:   format,
			}
			if arg {
				node.Positional = append(node.Positional, &value)
			} else {
				value.Flag = true
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

	return node
}
