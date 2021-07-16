package istio_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/istio"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/reference"
	"github.com/maistra/istio-workspace/test/testclient"
)

var _ = Describe("Operations for istio gateway kind", func() {

	var (
		objects      []runtime.Object
		c            client.Client
		ctx          model.SessionContext
		get          *testclient.Getters
		ref          model.Ref
		locators     model.LocatorStore
		modificators model.ModificatorStore
	)

	JustBeforeEach(func() {
		schema, _ := v1alpha1.SchemeBuilder.Register(
			&istionetwork.Gateway{},
			&istionetwork.GatewayList{}).Build()

		c = fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(objects...).Build()
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
					&istionetwork.Gateway{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "gateway-force-updated",
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
									},
								},
								{
									Port: &v1alpha3.Port{
										Protocol: "HTTP",
										Name:     "http",
										Number:   80,
									},
									Hosts: []string{
										"other-domain.com",
									},
								},
							},
						},
					},
				}
				ref = model.Ref{
					KindName: model.ParseRefKindName("customer-v1"),
				}
				locators = model.LocatorStore{}
				locators.Report(model.LocatorStatus{Resource: model.Resource{Kind: "Gateway", Namespace: "test", Name: "gateway"}, Action: model.ActionModify})
				modificators = model.ModificatorStore{}
			})

			It("should add reference", func() {
				istio.GatewayModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				gw := get.Gateway("test", "gateway")
				Expect(reference.Get(&gw)).To(HaveLen(1))
			})

			It("add single session", func() {
				istio.GatewayModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				gw := get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ConsistOf("domain.com", "test.domain.com"))
			})

			It("add single session - verify ref", func() {
				istio.GatewayModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				reported := modificators.Stored[0]
				Expect(reported.Name).To(Equal("gateway"))
				Expect(reported.Kind).To(Equal("Gateway"))
				Expect(reported.Prop["hosts"]).To(Equal("test.domain.com"))
			})

			It("add multiple session", func() {
				istio.GatewayModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				gw := get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ConsistOf("domain.com", "test.domain.com"))

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
				locators2 := model.LocatorStore{}
				locators2.Report(model.LocatorStatus{Resource: model.Resource{Kind: "Gateway", Namespace: "test", Name: "gateway"}, Action: model.ActionModify})
				modificators2 := model.ModificatorStore{}
				istio.GatewayModificator(ctx2, ref, locators2.Store, modificators2.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				gw = get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(3))
				Expect(gw.Spec.Servers[0].Hosts).To(ConsistOf("domain.com", "test.domain.com", "test2.domain.com"))
			})

			It("add multiple refs", func() {
				istio.GatewayModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				gw := get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ConsistOf("domain.com", "test.domain.com"))

				istio.GatewayModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(2))

				gw = get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ConsistOf("domain.com", "test.domain.com"))
			})

			It("should only return added hosts once", func() {
				locators.Clear()
				locators.Report(model.LocatorStatus{Resource: model.Resource{Kind: "Gateway", Namespace: "test", Name: "gateway-mutated"}, Action: model.ActionModify})

				istio.GatewayModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				status := modificators.Stored[0]
				Expect(status.Prop["hosts"]).To(Equal("test.domain.com"))
			})

			It("should reapply found ike hosts if gateway out of sync", func() {
				locators.Clear()
				locators.Report(model.LocatorStatus{Resource: model.Resource{Kind: "Gateway", Namespace: "test", Name: "gateway-force-updated"}, Action: model.ActionModify})

				istio.GatewayModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				gw := get.Gateway("test", "gateway-force-updated")
				Expect(gw.Spec.Servers[0].Hosts).To(ConsistOf("domain.com", "test.domain.com"))
				Expect(gw.Spec.Servers[1].Hosts).To(ConsistOf("other-domain.com", "test.other-domain.com"))
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
								{
									Port: &v1alpha3.Port{
										Protocol: "HTTP",
										Name:     "http",
										Number:   81,
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
				ref = model.Ref{
					KindName: model.ParseRefKindName("customer-v1"),
				}
				locators = model.LocatorStore{}
				locators.Report(model.LocatorStatus{Resource: model.Resource{Namespace: "test", Kind: "Gateway", Name: "gateway"}, Action: model.ActionRevert})
				modificators = model.ModificatorStore{}
			})

			It("remove reference", func() {
				istio.GatewayModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				gw := get.Gateway("test", "gateway")
				Expect(reference.Get(&gw)).To(BeEmpty())
			})

			It("single remove session", func() {
				istio.GatewayModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				gw := get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[1].Hosts).To(HaveLen(2))
			})

			It("multiple remove sessions", func() {
				istio.GatewayModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				gw := get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ConsistOf("domain.com", "test2.domain.com"))

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
				// Another Context would have had another Ref object with the Gateway Resource Status modified. simulate.
				locators2 := model.LocatorStore{}
				modificators2 := model.ModificatorStore{}
				locators2.Report(model.LocatorStatus{Resource: model.Resource{Kind: "Gateway", Namespace: "test", Name: "gateway"}, Action: model.ActionRevert})
				istio.GatewayModificator(ctx2, ref, locators2.Store, modificators2.Report)
				Expect(modificators2.Stored).To(HaveLen(1))

				gw = get.Gateway("test", "gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(1))
				Expect(gw.Spec.Servers[0].Hosts).To(ConsistOf("domain.com"))
				Expect(gw.Labels).ToNot(HaveKey(istio.LabelIkeHosts))
			})
		})
	})
})
