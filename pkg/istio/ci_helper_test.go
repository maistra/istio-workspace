package istio

import (
	"fmt"
	"testing"
)

// TODO: Tests disabled as they rely on oc and a current context setup outside of test. relying on oc is temp until istio API is ready.

func XTestDestinationRule(t *testing.T) {

	dr, err := getDestinationRuleMapped("bookinfo", "details")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(dr.Kind)
	fmt.Println(dr.Spec.Subsets[0].Labels)

	dr.Spec.Subsets[0].Name = "v1-test"
	setDestinationRule("bookinfo", dr)
}

func XTestVirtualService(t *testing.T) {

	vs, err := getVirtualServiceMapped("bookinfo", "details")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(vs.Spec.Http)

	vs.Spec.Http[0].WebsocketUpgrade = true
	setVirtualService("bookinfo", vs)
}
