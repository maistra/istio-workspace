package session

import (
	"fmt"
	"testing"

	istionetwork "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/networking/v1alpha3"
	"sigs.k8s.io/yaml"
)

func TestVirtualServiceMutatorAdd(t *testing.T) {
	vs := istionetwork.VirtualService{}
	err := yaml.Unmarshal([]byte(simpleVirtualService), &vs)
	if err != nil {
		t.Fatal(err)
	}

	vsm := VirtualServiceMutator{}
	vsa, err := vsm.Add(vs)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := yaml.Marshal(vsa)
	fmt.Println(string(spec))

	if len(vsa.Spec.Http) != 2 {
		t.Fatal("additional route not added")
	}

	if vsa.Spec.Http[0].Match == nil {
		t.Fatal("missing match route")
	}

	if vsa.Spec.Http[0].Route[0].Destination.Subset == "v1-test" {
		t.Fatal("missing subset update")
	}

}

func TestVirtualServiceMutatorRemove(t *testing.T) {
	vs := istionetwork.VirtualService{}
	err := yaml.Unmarshal([]byte(simpleMutatedVirtualService), &vs)
	if err != nil {
		t.Fatal(err)
	}

	vsm := VirtualServiceMutator{}
	vsa, err := vsm.Remove(vs)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := yaml.Marshal(vsa)
	fmt.Println(string(spec))

	if len(vsa.Spec.Http) > 1 {
		t.Fatal("route not removed")
	}

	if vsa.Spec.Http[0].Route[0].Destination.Subset != "v1" {
		t.Fatal("removed wrong destination")
	}
}

var simpleVirtualService = `kind: VirtualService
metadata:
  annotations:
  creationTimestamp: 2019-01-16T20:58:51Z
  generation: 1
  name: details
  namespace: bookinfo
  resourceVersion: "4978223"
  selfLink: /apis/networking.istio.io/v1alpha3/namespaces/bookinfo/virtualservices/details
  uid: 86e9c879-19d1-11e9-a489-482ae3045b54
spec:
  hosts:
  - details
  http:
  - route:
    - destination:
        host: details
        subset: v1
`
var simpleMutatedVirtualService = `kind: VirtualService
metadata:
  creationTimestamp: "2019-01-16T20:58:51Z"
  generation: 1
  name: details
  namespace: bookinfo
  resourceVersion: "4978223"
  selfLink: /apis/networking.istio.io/v1alpha3/namespaces/bookinfo/virtualservices/details
  uid: 86e9c879-19d1-11e9-a489-482ae3045b54
spec:
  hosts:
  - details
  http:
  - match:
    - headers:
        end-user:
          exact: jason
    route:
    - destination:
        host: details
        subset: v1-test
  - route:
    - destination:
        host: details
        subset: v1
`
