package session

import (
	"fmt"
	"strings"
	"testing"

	istionetwork "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/networking/v1alpha3"
	"sigs.k8s.io/yaml"
)

func TestDestinationRuleMutatorAdd(t *testing.T) {
	dr := istionetwork.DestinationRule{}
	err := yaml.Unmarshal([]byte(simpleDestinationRule), &dr)
	if err != nil {
		t.Fatal(err)
	}

	drm := DestinationRuleMutator{}
	dra, err := drm.Add(dr)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := yaml.Marshal(dra)
	fmt.Println(string(spec))

	if len(dra.Spec.Subsets) != 3 {
		t.Fatal("new subset not added")
	}

	if dra.Spec.Subsets[2].Name != "v1-test" {
		t.Fatal("wrong subset name")
	}
}

func TestDestinationRuleMutatorRemove(t *testing.T) {
	dr := istionetwork.DestinationRule{}
	err := yaml.Unmarshal([]byte(simpleMutatedDestinationRule), &dr)
	if err != nil {
		t.Fatal(err)
	}

	drm := DestinationRuleMutator{}
	dra, err := drm.Remove(dr)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := yaml.Marshal(dra)
	fmt.Println(string(spec))

	if len(dra.Spec.Subsets) != 2 {
		t.Fatal("new subset not added")
	}

	for _, s := range dra.Spec.Subsets {
		if strings.Contains(s.Name, "v1-test") {
			t.Fatal("wrong subset removed")
		}
	}
}

var simpleDestinationRule = `kind: DestinationRule
metadata:
  annotations:
  creationTimestamp: 2019-01-16T19:28:05Z
  generation: 1
  name: details
  namespace: bookinfo
  resourceVersion: "4955188"
  selfLink: /apis/networking.istio.io/v1alpha3/namespaces/bookinfo/destinationrules/details
  uid: d928001c-19c4-11e9-a489-482ae3045b54
spec:
  host: details
  subsets:
  - labels:
      version: v1
    name: v1
  - labels:
      version: v2
    name: v2
`

var simpleMutatedDestinationRule = `kind: DestinationRule
metadata:
  creationTimestamp: "2019-01-16T19:28:05Z"
  generation: 1
  name: details
  namespace: bookinfo
  resourceVersion: "4955188"
  selfLink: /apis/networking.istio.io/v1alpha3/namespaces/bookinfo/destinationrules/details
  uid: d928001c-19c4-11e9-a489-482ae3045b54
spec:
  host: details
  subsets:
  - labels:
      version: v1
    name: v1
  - labels:
      version: v2
    name: v2
  - labels:
      version: v1-test
    name: v1-test
`
