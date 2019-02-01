package apis

import (
	"github.com/aslakknutsen/istio-workspace/pkg/apis/istio/networking/v1alpha3"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha3.SchemeBuilder.AddToScheme)
}
