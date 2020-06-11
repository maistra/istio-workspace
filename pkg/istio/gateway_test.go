package istio

import (
	"github.com/maistra/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/test/operator"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Operations for istio gateway kind", func() {

	var (
		objects []runtime.Object
		c       client.Client
		ctx     model.SessionContext
		get     *operator.Helpers
		ref     *model.Ref
	)

	JustBeforeEach(func() {
		schema, _ := v1alpha1.SchemeBuilder.Register(
			&istionetwork.Gateway{},
			&istionetwork.GatewayList{}).Build()

		c = fake.NewFakeClientWithScheme(schema, objects...)
		get = operator.New(&c)
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
				}
				ref = &model.Ref{
					Name: "customer-v1",
					Targets: []model.LocatedResourceStatus{
						model.NewLocatedResource("Gateway", "gateway", nil),
					},
				}
			})

			It("add single session", func() {
				err := GatewayMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				gw := get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com", "test.domain.com"))
			})

			It("add multiple session", func() {
				err := GatewayMutator(ctx, ref)
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
				err = GatewayMutator(ctx2, ref)
				Expect(err).ToNot(HaveOccurred())

				gw = get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(3))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com", "test.domain.com", "test2.domain.com"))
			})

			It("add multiple refs", func() {
				err := GatewayMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				gw := get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com", "test.domain.com"))

				err = GatewayMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				gw = get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com", "test.domain.com"))
			})
		})

		Context("revertors", func() {

			BeforeEach(func() {
				objects = []runtime.Object{
					&istionetwork.Gateway{
						ObjectMeta: metav1.ObjectMeta{
							Name:        "gateway",
							Namespace:   "test",
							Annotations: map[string]string{LabelIkeHosts: "test.domain.com,test2.domain.com"},
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
						model.ResourceStatus{Kind: "Gateway", Name: "gateway", Action: model.ActionModified},
					},
				}
			})

			It("single remove session", func() {
				err := GatewayRevertor(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				gw := get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
			})

			It("multiple remove sessions", func() {
				err := GatewayRevertor(ctx, ref)
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
				err = GatewayRevertor(ctx2, ref)
				Expect(err).ToNot(HaveOccurred())

				gw = get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(1))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com"))
				Expect(gw.Labels).ToNot(HaveKey(LabelIkeHosts))
			})
		})
	})
})
