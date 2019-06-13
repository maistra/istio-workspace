package dynclient

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/maistra/istio-workspace/pkg/assets"

	openshiftApi "github.com/openshift/api/template/v1"

	"gopkg.in/yaml.v2"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
)

var re = regexp.MustCompile(`(?m)\${(.*?)}`)

const substitution = "{{ .$1 }}"

// ProcessTemplateUsingEnvVars takes template path, loads its content and
// substitutes variables by using those which are defined as environment variables.
// Returns processed template as byte array
func ProcessTemplateUsingEnvVars(templatePath string) ([]byte, error) {
	envMap := make(map[string]string)
	for _, v := range os.Environ() {
		envVar := strings.Split(v, "=")
		envMap[envVar[0]] = envVar[1]
	}

	return ProcessTemplate(templatePath, envMap)
}

// ProcessTemplate applies variables defined in the passed map to the
func ProcessTemplate(templatePath string, variables map[string]string) ([]byte, error) {
	data, err := Load(templatePath)
	if err != nil {
		return nil, err
	}

	var parameters struct {
		Parameter []openshiftApi.Parameter `yaml:"parameters"`
	}

	if err := yaml.Unmarshal(data, &parameters); err != nil {
		return nil, err
	}

	if len(variables) == 0 {
		variables = map[string]string{}
	}

	for _, v := range parameters.Parameter {
		if _, exists := variables[v.Name]; !exists {
			if v.Value == "" && v.Required {
				return nil, fmt.Errorf("expected %s to be defined but "+
					"can't be found in %s template nor as environment variable", v.Name, templatePath)
			}
			variables[v.Name] = v.Value
		}
	}

	tp := re.ReplaceAllString(string(data), substitution)
	t := template.Must(template.New(templatePath).Parse(tp))

	var processed bytes.Buffer
	if err := t.Execute(&processed, variables); err != nil {
		return nil, err
	}

	return processed.Bytes(), nil
}

// Load loads file asset into byte array
//
// Assets from given directory are added to the final binary through go-bindata code generation
func Load(filePath string) ([]byte, error) {
	data, err := assets.Asset(filePath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Parse takes byte array as a source and turns it into runtime.Object
func Parse(source []byte) (runtime.Object, error) {
	decode, err := Decoder()
	if err != nil {
		return nil, err
	}

	obj, _, err := decode(source, nil, nil)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// DecodeFunc is a function type matching Decoder function from k8s runtime apimachinery.
type DecodeFunc func(data []byte, defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error)

// Decoder registers required schemas containing objects to be deserialize from YAML
// and constructs decode function to be applied on the source YAML
func Decoder() (DecodeFunc, error) {
	s := runtime.NewScheme()
	if err := openshiftApi.AddToScheme(s); err != nil {
		return nil, err
	}
	if err := scheme.AddToScheme(s); err != nil {
		return nil, err
	}
	if err := apiextv1beta1.AddToScheme(s); err != nil {
		return nil, err
	}

	return serializer.NewCodecFactory(s).UniversalDeserializer().Decode, nil
}
