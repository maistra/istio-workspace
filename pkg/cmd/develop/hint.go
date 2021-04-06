package develop

import (
	"bytes"
	"text/template"

	istiov1alpha1 "github.com/maistra/istio-workspace/api/maistra/v1alpha1"
)

const urlHint = `Knowing your application url you can now access your new version by using
{{- $hosts := (findHosts .Ref) }}
{{ if $hosts }}
the following hosts
{{- range $hosts }}
$ curl {{ . }}
{{- end }}
{{ end -}}
{{- if .Route }}{{ if eq .Route.Type "header" }}
the following header
$ curl -H"{{.Route.Name}}:{{.Route.Value}}" YOUR_APP_URL.
{{ end }}{{ end }}
If you can't see any changes make sure that this header is respected by your app and propagated down the call chain.`

type data struct {
	Ref   *istiov1alpha1.RefStatus
	Route *istiov1alpha1.Route
}

// Hint returns a string containing the help for how to reach your new route.
func Hint(ref *istiov1alpha1.RefStatus, route *istiov1alpha1.Route) (string, error) {
	tmpl, err := getTemplate()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data{Ref: ref, Route: route})

	return buf.String(), err
}

func getTemplate() (*template.Template, error) {
	var err error
	t := template.New("workspace").Funcs(template.FuncMap{
		"findHosts": func(ref *istiov1alpha1.RefStatus) []string {
			if ref == nil {
				return []string{}
			}

			return ref.GetHostNames()
		},
	})
	t, err = t.Parse(urlHint)
	if err != nil {
		return nil, err
	}

	return t, nil
}
