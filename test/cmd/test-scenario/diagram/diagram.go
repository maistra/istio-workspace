package diagram

import (
	"fmt"
	"io"
	"strings"

	"github.com/blushft/go-diagrams/diagram"
	"github.com/blushft/go-diagrams/nodes/k8s"
	"github.com/maistra/istio-workspace/test/cmd/test-scenario/generator"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewPrinter(name string, out io.Writer) generator.Printer {
	objects := map[string]runtime.Object{}

	return func(object runtime.Object) {
		if obj, ok := meta.Accessor(object); ok == nil {

			fmt.Println(object.GetObjectKind().GroupVersionKind().Kind + ":" + obj.GetName())

			objects[object.GetObjectKind().GroupVersionKind().Kind+":"+obj.GetName()] = object

			// last object added
			if object.GetObjectKind().GroupVersionKind().Kind == "Gateway" && obj.GetName() == "test-gateway" {
				print(name, objects)
			}
		}
	}
}

func print(name string, objects map[string]runtime.Object) {
	d, err := diagram.New(diagram.Filename("test"), diagram.Label(name), diagram.Direction("LR"))
	if err != nil {
		fmt.Println(err)
	}

	for _, gw := range get(objects, byKind("Gateway")) {
		if g, ok := meta.Accessor(gw); ok == nil {
			gwn := k8s.Network.Ing(diagram.NodeLabel(g.GetName()))
			d.Add(gwn)
			for _, vs := range get(objects, byGatewayConnection(g.GetName())) {
				if v, ok := meta.Accessor(vs); ok == nil {
					vsn := k8s.Network.Svc(diagram.NodeLabel(v.GetName()))
					d.Connect(gwn, vsn)
				}
			}
		}
	}

	if err := d.Render(); err != nil {
		fmt.Println(err)
	}
}

type predicator func(key string, object runtime.Object) bool

func get(objects map[string]runtime.Object, pred predicator) []runtime.Object {
	found := []runtime.Object{}

	for key, val := range objects {
		if pred(key, val) {
			found = append(found, val)
		}
	}

	return found
}

func byKind(kind string) predicator {
	return func(key string, object runtime.Object) bool {
		return strings.HasPrefix(key, kind+":")
	}
}

func byGatewayConnection(gatewayName string) predicator {
	return func(key string, object runtime.Object) bool {
		found := false
		if obj, ok := object.(*istionetwork.VirtualService); ok {
			for _, gw := range obj.Spec.Gateways {
				if gw == gatewayName {
					found = true
				}
			}
		}
		return found
	}
}
