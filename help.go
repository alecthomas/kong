package kong

import (
	"io"
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

// WriteHelp to w.
//
// If w is nil, the default stdout writer will be used.
//
// If args are provided, help will be written in the context o
func (k *Kong) WriteHelp(w io.Writer, args ...interface{}) error {
	if w == nil {
		w = k.stdout
	}
	ctx := map[string]interface{}{
		"Application": k.Model,
	}
	for k, v := range k.helpContext {
		ctx[k] = v
	}
	return k.help.Execute(w, ctx)
}
