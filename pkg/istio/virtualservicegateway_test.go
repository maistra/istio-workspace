package istio_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	istionetworkv1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/istio"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model"
)

var _ = Describe("Location of Gateway connected VirtualService Kind", func() {

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
						Annotations: map[string]string{
							istio.LabelIkeHosts: "active-session.redhat-devcon.io",
						},
					},
					Spec: istionetworkv1alpha3.Gateway{
						Selector: map[string]string{"istio": "ingressgateway"},
						Servers: []*istionetworkv1alpha3.Server{
							{
								Port:  &istionetworkv1alpha3.Port{Name: "http", Protocol: "HTTP", Number: 80},
								Hosts: []string{"redhat-kubecon.io", "redhat-devcon.io", "active-session.redhat-devcon.io"},
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

			c = fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(objects...).Build()
			ctx = model.SessionContext{
				Name:      "test",
				Namespace: "bookinfo",
				Route:     model.Route{Type: "Header", Name: "x", Value: "y"},
				Client:    c,
				Log:       log.CreateOperatorAwareLogger("session").WithValues("type", "controller"),
			}
		})

		It("should expose hosts of located gateway", func() {
			ref := model.Ref{
				Name: "customer-v1",
			}

			found := istio.VirtualServiceGatewayLocator(ctx, &ref)
			Expect(found).To(BeTrue())

			gws := ref.GetTargets(model.Kind(istio.GatewayKind))
			Expect(gws).To(HaveLen(1))
			Expect(gws[0].Labels[istio.LabelIkeHosts]).ToNot(BeEmpty())
		})

		It("should only expose hosts not belonging to other sessions", func() {
			ref := model.Ref{
				Name: "customer-v1",
			}

			found := istio.VirtualServiceGatewayLocator(ctx, &ref)
			Expect(found).To(BeTrue())

			gws := ref.GetTargets(model.Kind(istio.GatewayKind))
			Expect(gws).To(HaveLen(1))
			Expect(gws[0].Labels[istio.LabelIkeHosts]).To(Equal("redhat-kubecon.io,redhat-devcon.io"))
		})
	})
})
