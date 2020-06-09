package istio

import (
	"github.com/maistra/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/model"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	istionetworkv1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var _ = Describe("Location of Gateway connected VirtualService Kind", func() {

	// Missing tests:
	//  - verify existing host logic from GW works (filter out our own sub domains)
	Context("gateway attachment", func() {

		var (
			objects []runtime.Object
			c       client.Client
			ctx     model.SessionContext
		)

		BeforeEach(func() {
			objects = []runtime.Object{
				&istionetwork.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "customer",
						Namespace: "bookinfo",
					},
					Spec: istionetworkv1alpha3.VirtualService{
						Hosts:    []string{"*"},
						Gateways: []string{"test-gateway"},
						Http: []*istionetworkv1alpha3.HTTPRoute{
							{
								Route: []*istionetworkv1alpha3.HTTPRouteDestination{
									{
										Destination: &istionetworkv1alpha3.Destination{
											Host: "customer",
											Port: &istionetworkv1alpha3.PortSelector{Number: 8080},
										},
									},
								},
							},
						},
					},
				},
				&istionetwork.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "non-customer",
						Namespace: "bookinfo",
					},
					Spec: istionetworkv1alpha3.VirtualService{
						Hosts:    []string{"*"},
						Gateways: []string{"test-gateway"},
						Http: []*istionetworkv1alpha3.HTTPRoute{
							{
								Route: []*istionetworkv1alpha3.HTTPRouteDestination{
									{
										Destination: &istionetworkv1alpha3.Destination{
											Host: "reviews",
											Port: &istionetworkv1alpha3.PortSelector{Number: 8080},
										},
									},
								},
							},
						},
					},
				},
				&istionetwork.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-gateway",
						Namespace: "bookinfo",
					},
					Spec: istionetworkv1alpha3.Gateway{
						Selector: map[string]string{"istio": "ingressgateway"},
						Servers: []*istionetworkv1alpha3.Server{
							{
								Port:  &istionetworkv1alpha3.Port{Name: "http", Protocol: "HTTP", Number: 80},
								Hosts: []string{"redhat-kubecon.io", "redhat-devcon.io"},
							},
						},
					},
				},
			}
		})

		JustBeforeEach(func() {
			schema, _ := v1alpha1.SchemeBuilder.Register(
				&istionetwork.VirtualServiceList{},
				&istionetwork.VirtualService{},
				&istionetwork.Gateway{}).Build()

			c = fake.NewFakeClientWithScheme(schema, objects...)
			ctx = model.SessionContext{
				Name:      "test",
				Namespace: "bookinfo",
				Route:     model.Route{Type: "Header", Name: "x", Value: "y"},
				Client:    c,
				Log:       logf.Log.WithName("controller_session"),
			}
		})

		It("should locate gateway", func() {
			ref := model.Ref{
				Name: "customer-v1",
			}

			found := VirtualServiceGatewayLocator(ctx, &ref)
			Expect(found).To(BeTrue())

			gws := ref.GetTargetsByKind(GatewayKind)
			Expect(gws).To(HaveLen(1))
			Expect(gws[0].Labels[LabelIkeHosts]).To(Equal("redhat-kubecon.io,redhat-devcon.io"))
		})
	})
})
