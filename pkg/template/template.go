package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	jsonpatch "github.com/evanphx/json-patch"
)

// NewDefaultEngine returns a new Engine with a predefined templates.
func NewDefaultEngine() *Engine {
	patches := []Patch{
		{
			Name: "prepared-image",
			Template: []byte(`[
					{{ template "_basic-version" . }}
				{{ if not (.Data.Has "/spec/template/spec/replicas") }}
				{"op": "add", "path": "/spec/template/spec/replicas", "value": {}},
				{{ end }}
				{"op": "replace", "path": "/spec/template/spec/replicas", "value": "1"},
				{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "{{.Vars.image}}"},

					{{ template "_basic-remove" . }}
				]`),
			Variables: map[string]string{
				"image": "",
			},
		},
		{
			Name: "telepresence",
			Template: []byte(`
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
`),
			Variables: map[string]string{
				"version": "",
			},
		},
		{
			Name: "_basic-version",
			Template: []byte(`
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
			`),
		},
		{
			Name: "_basic-remove",
			Template: []byte(`
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
			`),
		},
	}
	return NewEngine(patches)
}

// NewEngine constructs a new Engine with the given templates.
func NewEngine(patches Patches) *Engine {
	return &Engine{patches: patches}
}

// NewJSON constructs a JSON object from a json string.
func NewJSON(data []byte) (JSON, error) {
	t := JSON{}
	err := json.Unmarshal(data, &t)
	return t, err
}

// JSON is a parsed json structure with helper functions to access the data based on json paths.
type JSON map[string]interface{}

// Context contain the template context used during conversion. Holds template variables and data.
type Context struct {
	NewVersion string
	Data       JSON
	Vars       map[string]string
}

// Patch is a named JSON Patch and it's defined default variables.
type Patch struct {
	Name      string
	Template  []byte
	Variables map[string]string
}

// Patches holds all known patch templates for a Engine.
type Patches []Patch

// Engine is a reusable instance with a configured set of patch templates.
type Engine struct {
	patches Patches
}

// Value returns the object value behind a json path, e.g. /spec/metadata/name.
func (t JSON) Value(path string) (interface{}, error) {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("given path is not valid")
	}
	parts = parts[1:]
	var level interface{} = t
	for i, part := range parts {
		var l interface{}
		switch v := level.(type) {
		case map[string]interface{}:
			l = v[part]
		case JSON:
			l = v[part]
		case []interface{}:
			p, err := strconv.ParseInt(part, 10, 0)
			if err != nil {
				return nil, err
			}
			l = v[p]
		}

		switch v := l.(type) {
		case map[string]interface{}, []interface{}:
			level = v
		default:
			if i == len(parts)-1 {
				return l, nil
			}
			return nil, nil
		}
	}
	return level, nil
}

// Has is a check if the json contain a value behind a json path, e.g. /spec/metadata/name.
func (t JSON) Has(path string) bool {
	v, err := t.Value(path)
	if err != nil || v == nil {
		return false
	}
	return true
}

// Equal checks if the values are the same.
func (t JSON) Equal(path string, compare interface{}) bool {
	v, err := t.Value(path)
	if err != nil || v == nil {
		return false
	}

	return fmt.Sprint(v) == fmt.Sprint(compare)
}

// Run performs the template transformation of a given json structure.
func (e Engine) Run(name string, resource []byte, newVersion string, variables map[string]string) ([]byte, error) {
	t, err := loadTemplate(e.patches)
	if err != nil {
		return nil, err
	}

	patch := e.findPatch(name)
	if patch == nil {
		return nil, fmt.Errorf("unable to find patch %s", name)
	}

	patchVariables := map[string]string{}
	defaultVariables := patch.Variables

	for k, v := range defaultVariables {
		patchVariables[k] = v
	}
	for k, v := range variables {
		patchVariables[k] = v
	}

	resourceData, err := NewJSON(resource)
	if err != nil {
		return nil, err
	}

	c := Context{
		Data:       resourceData,
		NewVersion: newVersion,
		Vars:       patchVariables,
	}

	// Run Template
	rawPatch := new(bytes.Buffer)
	err = t.ExecuteTemplate(rawPatch, name, c)
	if err != nil {
		return nil, err
	}

	// Apply patch
	jsonPatch, err := jsonpatch.DecodePatch(rawPatch.Bytes())
	if err != nil {
		return nil, err
	}

	modified, err := jsonPatch.ApplyIndent(resource, "  ")
	if err != nil {
		return nil, err
	}

	return modified, nil
}

func (e Engine) findPatch(name string) *Patch {
	var patch *Patch
	for i, p := range e.patches {
		if p.Name == name {
			patch = &e.patches[i]
			break
		}
	}
	return patch
}

func loadTemplate(patches Patches) (*template.Template, error) {
	var err error
	t := template.New("workspace").Funcs(template.FuncMap{
		"failIfVariableDoesNotExist": func(vars map[string]string, name string) (string, error) {
			if vars == nil || vars[name] == "" {
				return "", fmt.Errorf("expected %s variable to be set", name)
			}
			return "", nil
		},
	})
	for _, p := range patches {
		t, err = t.New(p.Name).Parse(string(p.Template))
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
