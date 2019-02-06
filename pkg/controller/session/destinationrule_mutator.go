package session

import (
	"fmt"
	"strings"

	istionetwork "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/networking/v1alpha3"
	yaml "gopkg.in/yaml.v2"
	v1alpha3 "istio.io/api/networking/v1alpha3"
)

// DestinationRuleMutator mutates a DestinationRule by adding the required subset
type DestinationRuleMutator struct{}

// Add adds a mutation
func (d *DestinationRuleMutator) Add(dr istionetwork.DestinationRule) (istionetwork.DestinationRule, error) {

	{
		x, _ := yaml.Marshal(dr)
		fmt.Println(string(x))
	}

	dr.Spec.Subsets = append(dr.Spec.Subsets, &v1alpha3.Subset{
		Name: "v1-test",
		Labels: map[string]string{
			"version": "v1-test",
		},
	})

	{
		x, _ := yaml.Marshal(dr)
		fmt.Println(string(x))
	}

	return dr, nil
}

// Remove removes a mutation
func (d *DestinationRuleMutator) Remove(dr istionetwork.DestinationRule) (istionetwork.DestinationRule, error) {
	for i := 0; i < len(dr.Spec.Subsets); i++ {
		if strings.Contains(dr.Spec.Subsets[i].Name, "-test") {
			dr.Spec.Subsets = append(dr.Spec.Subsets[:i], dr.Spec.Subsets[i+1:]...)
			break
		}
	}
	return dr, nil
}
