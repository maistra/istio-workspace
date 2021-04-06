package api

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

// AddToSchemes may be used to add all resources defined in the project to a Scheme.
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme.
func AddToScheme(s *runtime.Scheme) error {
	return errors.Wrap(AddToSchemes.AddToScheme(s), "failed registering api")
}
