package generator

import (
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

// Printer is a function to output generated runtime.Objects.
type Printer func(object runtime.Object)

// WrapInYamlPrinter prints passed objects to io.Writer.
func WrapInYamlPrinter(out io.Writer) Printer {
	return func(object runtime.Object) {
		b, err := yaml.Marshal(object)
		if err != nil {
			panic(fmt.Sprintf("marshall error: %s\n", err.Error()))
		}
		if _, err = out.Write(b); err != nil {
			panic(fmt.Sprintf("failed writing object: %s\n", err.Error()))
		}
		if _, err = io.WriteString(out, "---\n"); err != nil {
			panic(fmt.Sprintf("failed writing delimiter: %s\n", err.Error()))
		}
	}
}
