{{ if not (.Data.Has "/spec/template/metadata") }}
{"op": "add", "path": "/spec/template/metadata", "value": {}},
{{ end }}
{{ if not (.Data.Has "/spec/template/metadata/labels") }}
{"op": "add", "path": "/spec/template/metadata/labels", "value": {}},
{{ end }}
{{ if .Data.Has "/spec/template/metadata/labels/version" }}
{"op": "copy", "from": "/spec/template/metadata/labels/version", "path": "/spec/template/metadata/labels/version-source"},
{"op": "replace", "path": "/spec/template/metadata/labels/version", "value": "{{.NewVersion}}"},
{{ end }}
{{ if not (.Data.Has "/spec/template/metadata/labels/version") }}
{"op": "add", "path": "/spec/template/metadata/labels/version", "value": "{{.NewVersion}}"},
{{ end }}
{{ if not (.Data.Has "/spec/selector") }}
{"op": "add", "path": "/spec/selector", "value": {}},
{{ end }}
{{ if .Data.Equal "/kind" "Deployment" }}
  {{ if not (.Data.Has "/spec/selector/matchLabels") }}
  {"op": "add", "path": "/spec/selector/matchLabels", "value": {}},
  {{ end }}
  {{ if .Data.Has "/spec/selector/matchLabels/version" }}
  {"op": "replace", "path": "/spec/selector/matchLabels/version", "value": "{{.NewVersion}}"},
  {{ end }}
  {{ if not (.Data.Has "/spec/selector/matchLabels/version") }}
  {"op": "add", "path": "/spec/selector/matchLabels/version", "value": "{{.NewVersion}}"},
  {{ end }}
{{ end }}
{{ if .Data.Equal "/kind" "DeploymentConfig" }}
  {{ if .Data.Has "/spec/selector/version" }}
  {"op": "replace", "path": "/spec/selector/version", "value": "{{.NewVersion}}"},
  {{ end }}
  {{ if not (.Data.Has "/spec/selector/version") }}
  {"op": "add", "path": "/spec/selector/version", "value": "{{.NewVersion}}"},
  {{ end }}
{{ end }}
{{ if .Data.Has "/metadata/labels/version" }}
{"op": "replace", "path": "/metadata/labels/version", "value": "{{.NewVersion}}"},
{{ end }}
{"op": "replace", "path": "/metadata/name", "value": "{{.Data.Value "/metadata/name"}}-{{.NewVersion}}"},
