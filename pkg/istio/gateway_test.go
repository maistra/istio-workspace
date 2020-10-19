package istio_test

import (
	"github.com/maistra/istio-workspace/pkg/apis/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/istio"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/test/testclient"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Operations for istio gateway kind", func() {

	var (
		objects []runtime.Object
		c       client.Client
		ctx     model.SessionContext
		get     *testclient.Getters
		ref     *model.Ref
	)

	JustBeforeEach(func() {
		schema, _ := v1alpha1.SchemeBuilder.Register(
			&istionetwork.Gateway{},
			&istionetwork.GatewayList{}).Build()

		c = fake.NewFakeClientWithScheme(schema, objects...)
		get = testclient.New(c)
		ctx = model.SessionContext{
			Name:      "test",
			Namespace: "test",
			Route:     model.Route{Type: "Header", Name: "x", Value: "y"},
			Client:    c,
			Log:       log.CreateOperatorAwareLogger("gateway"),
		}
	})

	Context("manipulation", func() {

		Context("mutators", func() {

			BeforeEach(func() {
				objects = []runtime.Object{
					&istionetwork.Gateway{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "gateway",
							Namespace: "test",
						},
						Spec: v1alpha3.Gateway{
							Selector: map[string]string{
								"istio": "ingressgateway",
							},
							Servers: []*v1alpha3.Server{
								{
									Port: &v1alpha3.Port{
										Protocol: "HTTP",
										Name:     "http",
										Number:   80,
									},
									Hosts: []string{
										"domain.com",
									},
								},
							},
						},
					},
					&istionetwork.Gateway{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "gateway-mutated",
							Namespace: "test",
							Annotations: map[string]string{
								istio.LabelIkeHosts: "test.domain.com",
							},
						},
						Spec: v1alpha3.Gateway{
							Selector: map[string]string{
								"istio": "ingressgateway",
							},
							Servers: []*v1alpha3.Server{
								{
									Port: &v1alpha3.Port{
										Protocol: "HTTP",
										Name:     "http",
										Number:   80,
									},
									Hosts: []string{
										"domain.com",
										"test.domain.com",
									},
								},
							},
						},
					},
				}
				ref = &model.Ref{
					Name: "customer-v1",
					Targets: []model.LocatedResourceStatus{
						model.NewLocatedResource("Gateway", "gateway", nil),
					},
				}
			})

			It("add single session", func() {
				err := istio.GatewayMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				gw := get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com", "test.domain.com"))
			})

			It("add single session - verify ref", func() {
				err := istio.GatewayMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				Expect(ref.ResourceStatuses).To(HaveLen(1))
				Expect(ref.ResourceStatuses[0].Name).To(Equal("gateway"))
				Expect(ref.ResourceStatuses[0].Kind).To(Equal("Gateway"))
				Expect(ref.ResourceStatuses[0].Prop["hosts"]).To(Equal("test.domain.com"))
			})

			It("add multiple session", func() {
				err := istio.GatewayMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				gw := get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com", "test.domain.com"))

				ctx2 := model.SessionContext{
					Name:      "test2",
					Namespace: "test",
					Route: model.Route{
						Type:  "header",
						Name:  "test",
						Value: "x",
					},
					Client: ctx.Client,
					Log:    ctx.Log,
				}
				err = istio.GatewayMutator(ctx2, ref)
				Expect(err).ToNot(HaveOccurred())

				gw = get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(3))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com", "test.domain.com", "test2.domain.com"))
			})

			It("add multiple refs", func() {
				err := istio.GatewayMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				gw := get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com", "test.domain.com"))

				err = istio.GatewayMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				gw = get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com", "test.domain.com"))
			})

			It("should only return added hosts once", func() {
				ref.Targets = []model.LocatedResourceStatus{
					model.NewLocatedResource("Gateway", "gateway-mutated", nil),
				}

				err := istio.GatewayMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				statuss := ref.GetResources(model.Kind(istio.GatewayKind))
				Expect(statuss).To(HaveLen(1))

				status := statuss[0]
				Expect(status.Prop["hosts"]).To(Equal("test.domain.com"))
			})
		})

		Context("revertors", func() {

			BeforeEach(func() {
				objects = []runtime.Object{
					&istionetwork.Gateway{
						ObjectMeta: metav1.ObjectMeta{
							Name:        "gateway",
							Namespace:   "test",
							Annotations: map[string]string{istio.LabelIkeHosts: "test.domain.com,test2.domain.com"},
						},
						Spec: v1alpha3.Gateway{
							Selector: map[string]string{
								"istio": "ingressgateway",
							},
							Servers: []*v1alpha3.Server{
								{
									Port: &v1alpha3.Port{
										Protocol: "HTTP",
										Name:     "http",
										Number:   80,
									},
									Hosts: []string{
										"domain.com",
										"test.domain.com",
										"test2.domain.com",
									},
								},
							},
						},
					},
				}
				ref = &model.Ref{
					Name: "customer-v1",
					ResourceStatuses: []model.ResourceStatus{
						{Kind: "Gateway", Name: "gateway", Action: model.ActionModified},
					},
				}
			})

			It("single remove session", func() {
				err := istio.GatewayRevertor(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				gw := get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
			})

			It("multiple remove sessions", func() {
				err := istio.GatewayRevertor(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				gw := get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com", "test2.domain.com"))

				ctx2 := model.SessionContext{
					Name:      "test2",
					Namespace: "test",
					Route: model.Route{
						Type:  "header",
						Name:  "test",
						Value: "x",
					},
					Client: ctx.Client,
					Log:    ctx.Log,
				}
				// Another Context would have had another Ref object with the Gateway Resource Status modifed. simulate.
				ref.AddResourceStatus(model.ResourceStatus{Kind: "Gateway", Name: "gateway", Action: model.ActionModified})
				err = istio.GatewayRevertor(ctx2, ref)
				Expect(err).ToNot(HaveOccurred())

				gw = get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(1))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com"))
				Expect(gw.Labels).ToNot(HaveKey(istio.LabelIkeHosts))
			})
		})
	})
})
