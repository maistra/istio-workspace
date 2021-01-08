{{ failIfVariableDoesNotExist .Vars "version" -}}

[
  {{ template "_basic-version" . }}

  {{ if not (.Data.Has "/spec/template/spec/replicas") }}
  {"op": "add", "path": "/spec/template/spec/replicas", "value": {}},
  {{ end }}
  {"op": "replace", "path": "/spec/template/spec/replicas", "value": "1"},
  {"op": "add", "path": "/spec/template/metadata/labels/telepresence", "value": "test"},
  {"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "datawire/telepresence-k8s:{{.Vars.version}}"},
  {{ if not (.Data.Has "/spec/template/spec/containers/0/env") }}
  {"op": "add", "path": "/spec/template/spec/containers/0/env", "value": []},
  {{ end }}
  {"op": "add", "path": "/spec/template/spec/containers/0/env/-", "value": {
    "name": "TELEPRESENCE_CONTAINER_NAMESPACE",
    "valueFrom": {
      "fieldRef": {
        "apiVersion": "v1",
        "fieldPath": "metadata.namespace"
      }
    }
  }
  },
  {{ if .Data.Has "/spec/template/spec/containers/0/args" }}
  {"op": "remove", "path": "/spec/template/spec/containers/0/args"},
  {{ end }}
  {{ if .Data.Has "/spec/template/spec/containers/0/command" }}
  {"op": "remove", "path": "/spec/template/spec/containers/0/command"},
  {{ end }}

    {{ template "_basic-remove" . }}
]
