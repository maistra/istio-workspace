{{ if .Data.Has "/spec/template/spec/containers/0/livenessProbe" }}
{"op": "remove", "path": "/spec/template/spec/containers/0/livenessProbe"},
{{ end }}
{{ if .Data.Has "/spec/template/spec/containers/0/readinessProbe" }}
{"op": "remove", "path": "/spec/template/spec/containers/0/readinessProbe"},
{{ end }}
{{ if .Data.Has "/metadata/resourceVersion" }}
{"op": "remove", "path": "/metadata/resourceVersion"},
{{ end }}
{{ if .Data.Has "/metadata/generation" }}
{"op": "remove", "path": "/metadata/generation"},
{{ end }}
{{ if .Data.Has "/metadata/uid" }}
{"op": "remove", "path": "/metadata/uid"},
{{ end }}
{{ if .Data.Has "/metadata/creationTimestamp" }}
{"op": "remove", "path": "/metadata/creationTimestamp"}
{{ end }}
