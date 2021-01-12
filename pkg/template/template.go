package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/maistra/istio-workspace/pkg/assets"

	jsonpatch "github.com/evanphx/json-patch"
)

const TemplatePath = "TEMPLATE_PATH"

func loadPatches(tplFolder string) []Patch {
	tplDir, err := assets.ListDir(tplFolder)
	if err != nil {
		panic(err)
	}
	patches := []Patch{}
	for _, file := range tplDir {
		if !strings.HasSuffix(file, ".tpl") {
			continue
		}
		tplName := strings.Replace(file, ".tpl", "", -1)
		tpl, err := assets.Load(tplFolder + "/" + file)
		if err != nil {
			panic(err)
		}
		tplVars := map[string]string{}
		if tplVarRaw, err := assets.Load(tplFolder + "/" + tplName + ".var"); err == nil {
			for _, line := range strings.Split(string(tplVarRaw), "\n") {
				if line != "" {
					vars := strings.Split(line, "=")
					varName := strings.Trim(vars[0], " ")
					tplVars[varName] = ""
					if len(vars) == 2 {
						tplVars[varName] = strings.Trim(vars[1], " ")
					}
				}
			}
		}
		patches = append(patches, Patch{
			Name:      tplName,
			Template:  tpl,
			Variables: tplVars,
		})
	}

	return patches
}

// NewDefaultEngine returns a new Engine with a predefined templates.
func NewDefaultEngine() Engine {
	return NewDefaultPatchEngine("template/strategies")
}

// NewDefaultPatchEngine returns a new Engine with a predefined templates.
func NewDefaultPatchEngine(path string) Engine {
	return NewPatchEngine(loadPatches(path))
}

// NewPatchEngine constructs a new Engine with the given templates.
func NewPatchEngine(patches Patches) Engine {
	return &patchEngine{patches: patches}
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

// Engine is a interface that describes a way to prepare the Deployment for cloning.
type Engine interface {
	Run(name string, resource []byte, newVersion string, variables map[string]string) ([]byte, error)
}

// PatchEngine is a reusable instance with a configured set of patch templates to manipulate the Deployment object via json patches.
type patchEngine struct {
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
func (e patchEngine) Run(name string, resource []byte, newVersion string, variables map[string]string) ([]byte, error) {
	t, err := parseTemplate(e.patches)
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

func (e patchEngine) findPatch(name string) *Patch {
	var patch *Patch
	for i, p := range e.patches {
		if p.Name == name {
			patch = &e.patches[i]
			break
		}
	}
	return patch
}

func parseTemplate(patches Patches) (*template.Template, error) {
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
