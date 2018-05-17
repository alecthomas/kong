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
		if test, ok := recover().(error); ok {
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

	return buildNode(iv), nil
}

func buildNode(v reflect.Value) *Node {
	node := &Node{}
	for i := 0; i < v.NumField(); i++ {
		ft := v.Type().Field(i)
		fv := v.Field(i)

		name := ft.Tag.Get("name")
		if name == "" {
			name = strings.ToLower(strings.Join(camelCase(ft.Name), "-"))
		}
		help := ft.Tag.Get("help")
		decoder := DecoderForField(ft)
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
		_, arg := ft.Tag.Lookup("arg")
		env := ft.Tag.Get("env")

		// Nested structs are commands.
		if ft.Type.Kind() == reflect.Struct {
			child := buildNode(fv)
			child.Help = help

			// A branching argument. This is a bit hairy, as we let buildNode() do the parsing, then check that
			// a positional argument is provided to the child, and move it to the branching argument field.
			if arg {
				if len(child.Positional) == 0 {
					panic(fmt.Errorf("positional branch %s.%s must have at least one child positional argument",
						v.Type().Name(), ft.Name))
				}
				value := child.Positional[0]
				child.Positional = child.Positional[1:]
				if child.Help == "" {
					child.Help = value.Help
				}
				child.Name = value.Name
				node.Children = append(node.Children, &Branch{Argument: &Argument{
					Node:     *child,
					Argument: value,
				}})
			} else {
				child.Name = name
				node.Children = append(node.Children, &Branch{Command: child})
			}
		} else {
			value := Value{
				Name:     name,
				Help:     help,
				Decoder:  decoder,
				Value:    fv,
				Field:    ft,
				Required: !optional || required,
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

	return node
}
