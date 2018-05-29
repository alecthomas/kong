package kong

import (
	"bytes"
	"fmt"
	"go/doc"
	"io"
	"reflect"
	"strings"

	"github.com/aymerick/raymond"
)

const (
	defaultIndent   = 2
	defaultTemplate = `
{{#with App}}
usage: {{Name}}

{{#wrap}}
{{Help}}
{{/wrap}}

Flags:
{{#indent}}
{{formatFlags Flags}}
{{/indent}}

{{#if Children}}
{{#indent}}
{{#each Children}}
{{Name}}
{{/each}}
{{/indent}}
{{/if}}
{{/with}}
`
)

var defaultHelpTemplate = raymond.MustParse(strings.TrimSpace(defaultTemplate))

func init() {
	defaultHelpTemplate.RegisterHelpers(map[string]interface{}{
		"indent": func(options *raymond.Options) string {
			indent, ok := options.HashProp("depth").(int)
			if !ok {
				indent = 2
			}
			width := options.Data("width").(int)
			frame := options.NewDataFrame()
			frame.Set("width", width-indent)
			indentStr := strings.Repeat(" ", indent)
			lines := strings.Split(options.FnData(frame), "\n")
			for i, line := range lines {
				lines[i] = indentStr + line
			}
			return strings.Join(lines, "\n")
		},
		"formatFlags": func(flags []*Flag, options *raymond.Options) string {
			rows := [][2]string{}
			haveShort := false
			for _, flag := range flags {
				if flag.Short != 0 {
					haveShort = true
					break
				}
			}
			for _, flag := range flags {
				if !flag.Hidden {
					rows = append(rows, [2]string{formatFlag(haveShort, flag), flag.Help})
				}
			}
			w := bytes.NewBuffer(nil)
			formatTwoColumns(w, 0, 2, options.Data("width").(int), rows)
			return w.String()
		},
		"wrap": func(options *raymond.Options) string {
			w := bytes.NewBuffer(nil)
			doc.ToText(w, options.Fn(), "", "  ", options.Data("width").(int))
			return w.String()
		},
	})
}

// Help returns a Hook that will display help and exit.
//
// tmpl receives a context with several top-level values, in addition to those passed through tmplctx:
// .Context which is of type *Context and .Path which is of type *Path.
func Help(tmpl *raymond.Template, tmplctx map[string]interface{}) HookFunction {
	return func(ctx *Context, path *Path) error {
		merged := map[string]interface{}{
			"App":     ctx.App,
			"Context": ctx,
			"Path":    path,
		}
		for k, v := range tmplctx {
			merged[k] = v
		}
		frame := raymond.NewDataFrame()
		frame.Set("width", guessWidth(ctx.App.Stdout))
		output, err := tmpl.ExecWith(merged, frame)
		if err != nil {
			return err
		}
		io.WriteString(ctx.App.Stdout, output)
		ctx.App.Exit(0)
		return nil
	}
}

func formatTwoColumns(w io.Writer, indent, padding, width int, rows [][2]string) {
	// Find size of first column.
	s := 0
	for _, row := range rows {
		if c := len(row[0]); c > s && c < 30 {
			s = c
		}
	}

	indentStr := strings.Repeat(" ", indent)
	offsetStr := strings.Repeat(" ", s+padding)

	for _, row := range rows {
		buf := bytes.NewBuffer(nil)
		doc.ToText(buf, row[1], "", strings.Repeat(" ", defaultIndent), width-s-padding-indent)
		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		fmt.Fprintf(w, "%s%-*s%*s", indentStr, s, row[0], padding, "")
		if len(row[0]) >= 30 {
			fmt.Fprintf(w, "\n%s%s", indentStr, offsetStr)
		}
		fmt.Fprintf(w, "%s\n", lines[0])
		for _, line := range lines[1:] {
			fmt.Fprintf(w, "%s%s%s\n", indentStr, offsetStr, line)
		}
	}
}

// haveShort will be true if there are short flags present at all in the help. Useful for column alignment.
func formatFlag(haveShort bool, flag *Flag) string {
	flagString := ""
	name := flag.Name
	isBool := flag.IsBool()
	if flag.Short != 0 {
		flagString += fmt.Sprintf("-%c, --%s", flag.Short, name)
	} else {
		if haveShort {
			flagString += fmt.Sprintf("    --%s", name)
		} else {
			flagString += fmt.Sprintf("--%s", name)
		}
	}
	if !isBool {
		flagString += fmt.Sprintf("=%s", flag.FormatPlaceHolder())
	}
	if flag.Value.Value.Kind() == reflect.Slice {
		flagString += " ..."
	}
	return flagString
}
