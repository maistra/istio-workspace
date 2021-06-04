package develop

import (
	"bytes"
	"text/template"

	"emperror.dev/errors"

	istiov1alpha1 "github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/internal/session"
)

const urlHint = `Knowing your application url you can now access your new version by using
{{- if .Hosts }}
the following hosts
{{- range .Hosts }}
$ curl {{ . }}
{{- end }}
{{ end -}}
{{- if .Route }}{{ if eq .Route.Type "header" }}
the following header
$ curl -H"{{.Route.Name}}:{{.Route.Value}}" YOUR_APP_URL.
{{ end }}{{ end }}
If you can't see any changes make sure that this header is respected by your app and propagated down the call chain.`

type data struct {
	Hosts []string
	Route *istiov1alpha1.Route
}

// Hint returns a string containing the help for how to reach your new route.
func Hint(state *session.State) (string, error) {
	tmpl, err := getTemplate()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data{Hosts: state.Hosts, Route: &state.Route})

	return buf.String(), err
}

func getTemplate() (*template.Template, error) {
	t, err := template.New("workspace").Parse(urlHint)

	return t, errors.Wrap(err, "failed parsing template")
}
