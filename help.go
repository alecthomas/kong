package kong

import (
	"text/template"
)

const defaultHelp = `{{- with .Application -}}
usage: {{.Name}}

{{.Help}}
{{range .Context.Flags}}
--{{.Name}}
{{end}}

{{- end -}}
`

var defaultHelpTemplate = template.Must(template.New("help").Parse(defaultHelp))

// Help returns a Hook that will display help and exit.
func Help(tmpl *template.Template, tmplctx map[string]interface{}) HookFunction {
	return func(app *Kong, ctx *Context, trace *Trace) error {
		merged := map[string]interface{}{
			"Application": app.Model,
		}
		for k, v := range tmplctx {
			merged[k] = v
		}
		err := tmpl.Execute(app.Stdout, merged)
		if err != nil {
			return err
		}
		app.Exit(0)
		return nil
	}
}
