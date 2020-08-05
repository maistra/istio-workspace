package main

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/maistra/istio-workspace/pkg/openshift/parser"
	openshiftApi "github.com/openshift/api/template/v1"
	"os"
)

// Process openshift template provided as path in the first argument.
// Prints processed template evaluating env vars to stdout or fails in case of error.
func main() {
	tplFile := os.Args[1]

	content, err := parser.ProcessTemplateUsingEnvVars(tplFile)
	if err != nil {
		panic(err)
	}

	tpl, err := parser.Parse(content)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	r := tpl.(*openshiftApi.Template)
	for _, obj := range r.Objects {
		yml, err := yaml.Marshal(obj)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(string(yml))
		fmt.Println("---")
	}

}
