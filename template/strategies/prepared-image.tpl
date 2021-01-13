[

  {{ template "_basic-version" . }}

  {{ if not (.Data.Has "/spec/template/spec/replicas") }}
  {"op": "add", "path": "/spec/template/spec/replicas", "value": {}},
  {{ end }}
  {"op": "replace", "path": "/spec/template/spec/replicas", "value": "1"},
  {"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "{{.Vars.image}}"},

  {{ template "_basic-remove" . }}
]
