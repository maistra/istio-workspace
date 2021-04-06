package istio //nolint:testpackage //reason we want to test mutationRequired in isolation

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
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/reference"
	"github.com/maistra/istio-workspace/test/testclient"
)

var _ = Describe("Operations for istio VirtualService kind", func() {

	var (
		objects []runtime.Object
		c       client.Client
		ctx     model.SessionContext
		get     *testclient.Getters
	)

	JustBeforeEach(func() {
		schema, _ := v1alpha1.SchemeBuilder.Register(
			&istionetwork.VirtualService{},
			&istionetwork.VirtualServiceList{},
			&istionetwork.Gateway{}).Build()

		c = fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(objects...).Build()
		get = testclient.New(c)
		ctx = model.SessionContext{
			Name:      "vs-test",
			Namespace: "test",
			Route:     model.Route{Type: "header", Name: "test", Value: "x"},
			Client:    c,
			Log:       log.CreateOperatorAwareLogger("destinationrule"),
		}
	})

	Context("manipulation", func() {

		Context("mutators", func() {

			BeforeEach(func() {
				objects = []runtime.Object{
					&istionetwork.VirtualService{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "details",
							Namespace: "test",
						},
						Spec: istionetworkv1alpha3.VirtualService{
							Hosts: []string{"details"},
							Http: []*istionetworkv1alpha3.HTTPRoute{
								{
									Route: []*istionetworkv1alpha3.HTTPRouteDestination{
										{
											Destination: &istionetworkv1alpha3.Destination{
												Host:   "details",
												Subset: "v1",
											},
											Weight: 50,
										},
										{
											Destination: &istionetworkv1alpha3.Destination{
												Host:   "details",
												Subset: "v2",
											},
											Weight: 50,
										},
									},
									Match: []*istionetworkv1alpha3.HTTPMatchRequest{
										{
											Uri: &istionetworkv1alpha3.StringMatch{MatchType: &istionetworkv1alpha3.StringMatch_Prefix{Prefix: "/a"}},
										},
										{
											Uri: &istionetworkv1alpha3.StringMatch{MatchType: &istionetworkv1alpha3.StringMatch_Prefix{Prefix: "/b"}},
										},
									},
									Mirror: &istionetworkv1alpha3.Destination{
										Host:   "details",
										Subset: "v3",
									},
									Redirect: &istionetworkv1alpha3.HTTPRedirect{
										Uri: "/redirected",
									},
								},
								{
									Route: []*istionetworkv1alpha3.HTTPRouteDestination{
										{
											Destination: &istionetworkv1alpha3.Destination{
												Host:   "details",
												Subset: "v4",
											},
										},
									},
								},
								{
									Route: []*istionetworkv1alpha3.HTTPRouteDestination{
										{
											Destination: &istionetworkv1alpha3.Destination{
												Host:   "details",
												Subset: "v5",
											},
										},
									},
									Match: []*istionetworkv1alpha3.HTTPMatchRequest{
										{
											Headers: map[string]*istionetworkv1alpha3.StringMatch{
												"request-id": {MatchType: &istionetworkv1alpha3.StringMatch_Exact{Exact: "test"}},
											},
										},
									},
								},
								{
									Route: []*istionetworkv1alpha3.HTTPRouteDestination{
										{
											Destination: &istionetworkv1alpha3.Destination{
												Host: "x",
											},
										},
									},
								},
							},
						},
					},
				}
			})

			Context("existing rule", func() {
				GetMutatedRoute := func(vs istionetwork.VirtualService, host model.HostName, subset string) *istionetworkv1alpha3.HTTPRoute {
					for _, h := range vs.Spec.Http {
						for _, r := range h.Route {
							if host.Match(r.Destination.Host) && r.Destination.Subset == subset {
								return h
							}
						}
					}
					return nil
				}
				var (
					ref            model.Ref
					targetV1       = model.NewLocatedResource("Deployment", "details-v1", map[string]string{"version": "v1"})
					targetV1Host   = model.HostName{Name: "details"}
					targetV1Subset = model.GetSha("v1") + "-vs-test"
					targetV4       = model.NewLocatedResource("Deployment", "details-v4", map[string]string{"version": "v4"})
					targetV4Host   = model.HostName{Name: "details"}
					targetV4Subset = model.GetSha("v4") + "-vs-test"
					targetV5       = model.NewLocatedResource("Deployment", "details-v5", map[string]string{"version": "v5"})
					targetV5Host   = model.HostName{Name: "details"}
					targetV5Subset = model.GetSha("v5") + "-vs-test"
					targetV6       = model.NewLocatedResource("Deployment", "x-v5", map[string]string{"version": "v5"})
					targetV6Host   = model.HostName{Name: "x"}
					targetV6Subset = model.GetSha("v5") + "-vs-test"
				)

				BeforeEach(func() {
					ref = model.Ref{
						Name:      "details-v1",
						Namespace: "test",
					}
				})

				It("should add reference", func() {
					ref.AddTargetResource(targetV1)
					ref.AddTargetResource(model.NewLocatedResource("Service", "details", map[string]string{}))

					err := VirtualServiceMutator(ctx, &ref)
					Expect(err).ToNot(HaveOccurred())

					virtualService := get.VirtualService("test", "details")
					Expect(reference.Get(&virtualService)).To(HaveLen(1))
				})

				It("route added", func() {
					ref.AddTargetResource(targetV1)
					ref.AddTargetResource(model.NewLocatedResource("Service", "details", map[string]string{}))

					err := VirtualServiceMutator(ctx, &ref)
					Expect(err).ToNot(HaveOccurred())

					virtualService := get.VirtualService("test", "details")
					Expect(GetMutatedRoute(virtualService, targetV1Host, targetV1Subset)).ToNot(BeNil())
				})

				It("route added before target route", func() {
					ref.AddTargetResource(targetV1)
					ref.AddTargetResource(model.NewLocatedResource("Service", "details", map[string]string{}))

					err := VirtualServiceMutator(ctx, &ref)
					Expect(err).ToNot(HaveOccurred())

					virtualService := get.VirtualService("test", "details")
					Expect(virtualService.Spec.Http[0].Route[0].Destination.Subset).To(Equal(targetV1Subset))
				})

				It("has match", func() {
					ref.AddTargetResource(targetV1)
					ref.AddTargetResource(model.NewLocatedResource("Service", "details", map[string]string{}))

					err := VirtualServiceMutator(ctx, &ref)
					Expect(err).ToNot(HaveOccurred())

					virtualService := get.VirtualService("test", "details")
					Expect(GetMutatedRoute(virtualService, targetV1Host, targetV1Subset).Match).ToNot(BeNil())
				})

				It("has subset", func() { // covered by GetMutatedRoute
					ref.AddTargetResource(targetV1)
					ref.AddTargetResource(model.NewLocatedResource("Service", "details", map[string]string{}))

					err := VirtualServiceMutator(ctx, &ref)
					Expect(err).ToNot(HaveOccurred())

					virtualService := get.VirtualService("test", "details")
					Expect(GetMutatedRoute(virtualService, targetV1Host, targetV1Subset).Route[0].Destination.Subset).To(Equal(targetV1Subset))
				})

				It("create match when no match found", func() {
					ref.AddTargetResource(targetV4)
					ref.AddTargetResource(model.NewLocatedResource("Service", "details", map[string]string{}))

					err := VirtualServiceMutator(ctx, &ref)
					Expect(err).ToNot(HaveOccurred())

					virtualService := get.VirtualService("test", "details")

					mutated := GetMutatedRoute(virtualService, targetV4Host, targetV4Subset)
					Expect(mutated).ToNot(BeNil())
					Expect(mutated.Match).To(HaveLen(1))
					for _, m := range mutated.Match {
						Expect(m.Headers).To(HaveLen(1))
						Expect(m.Headers["test"].GetExact()).To(Equal("x"))
					}
				})

				It("add route headers to found match with no headers", func() {
					ref.AddTargetResource(targetV1)
					ref.AddTargetResource(model.NewLocatedResource("Service", "details", map[string]string{}))

					err := VirtualServiceMutator(ctx, &ref)
					Expect(err).ToNot(HaveOccurred())

					virtualService := get.VirtualService("test", "details")

					mutated := GetMutatedRoute(virtualService, targetV1Host, targetV1Subset)
					Expect(mutated).ToNot(BeNil())
					Expect(mutated.Match).To(HaveLen(2))
					for _, m := range mutated.Match {
						Expect(m.Headers).To(HaveLen(1))
						Expect(m.Headers["test"].GetExact()).To(Equal("x"))
					}
				})

				It("add route headers to found match with found headers", func() {
					ref.AddTargetResource(targetV5)
					ref.AddTargetResource(model.NewLocatedResource("Service", "details", map[string]string{}))

					err := VirtualServiceMutator(ctx, &ref)
					Expect(err).ToNot(HaveOccurred())

					virtualService := get.VirtualService("test", "details")

					mutated := GetMutatedRoute(virtualService, targetV5Host, targetV5Subset)
					Expect(mutated).ToNot(BeNil())
					Expect(mutated.Match).To(HaveLen(1))
					for _, m := range mutated.Match {
						Expect(m.Headers).To(HaveLen(2))
						Expect(m.Headers["test"].GetExact()).To(Equal("x"))
					}
				})

				It("remove weighted destination", func() {
					ref.AddTargetResource(targetV1)
					ref.AddTargetResource(model.NewLocatedResource("Service", "details", map[string]string{}))

					err := VirtualServiceMutator(ctx, &ref)
					Expect(err).ToNot(HaveOccurred())

					virtualService := get.VirtualService("test", "details")
					Expect(GetMutatedRoute(virtualService, targetV1Host, targetV1Subset).Route[0].Weight).To(Equal(int32(0)))
				})

				It("remove other destinations", func() {
					ref.AddTargetResource(targetV1)
					ref.AddTargetResource(model.NewLocatedResource("Service", "details", map[string]string{}))

					err := VirtualServiceMutator(ctx, &ref)
					Expect(err).ToNot(HaveOccurred())

					virtualService := get.VirtualService("test", "details")
					Expect(GetMutatedRoute(virtualService, targetV1Host, targetV1Subset).Route).To(HaveLen(1))
				})

				It("remove mirror", func() {
					ref.AddTargetResource(targetV1)
					ref.AddTargetResource(model.NewLocatedResource("Service", "details", map[string]string{}))

					err := VirtualServiceMutator(ctx, &ref)
					Expect(err).ToNot(HaveOccurred())

					virtualService := get.VirtualService("test", "details")
					Expect(GetMutatedRoute(virtualService, targetV1Host, targetV1Subset).Mirror).To(BeNil())
				})

				It("remove redirect", func() {
					ref.AddTargetResource(targetV1)
					ref.AddTargetResource(model.NewLocatedResource("Service", "details", map[string]string{}))

					err := VirtualServiceMutator(ctx, &ref)
					Expect(err).ToNot(HaveOccurred())

					virtualService := get.VirtualService("test", "details")
					Expect(GetMutatedRoute(virtualService, targetV1Host, targetV1Subset).Redirect).To(BeNil())
				})

				It("include destinations with no subset", func() {
					ref.AddTargetResource(targetV6)
					ref.Name = "x-v5"
					ref.AddTargetResource(model.NewLocatedResource("Service", "x", map[string]string{}))

					err := VirtualServiceMutator(ctx, &ref)
					Expect(err).ToNot(HaveOccurred())

					virtualService := get.VirtualService("test", "details")
					Expect(GetMutatedRoute(virtualService, targetV6Host, targetV6Subset).Redirect).To(BeNil())
				})
			})
		})

		Context("revertors", func() {

			BeforeEach(func() {
				objects = []runtime.Object{
					&istionetwork.VirtualService{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "details",
							Namespace: "test",
						},
						Spec: istionetworkv1alpha3.VirtualService{
							Hosts: []string{"details"},
							Http: []*istionetworkv1alpha3.HTTPRoute{
								{
									Route: []*istionetworkv1alpha3.HTTPRouteDestination{
										{
											Destination: &istionetworkv1alpha3.Destination{
												Host:   "details",
												Subset: model.GetSha("v1") + "-vs-test",
											},
										},
									},
									Match: []*istionetworkv1alpha3.HTTPMatchRequest{
										{
											Headers: map[string]*istionetworkv1alpha3.StringMatch{
												"x-test-suite": {MatchType: &istionetworkv1alpha3.StringMatch_Exact{Exact: "feature-x"}},
											},
										},
									},
								},
								{
									Route: []*istionetworkv1alpha3.HTTPRouteDestination{
										{
											Destination: &istionetworkv1alpha3.Destination{
												Host:   "details",
												Subset: model.GetSha("v1") + "-vs-test",
											},
										},
									},
									Match: []*istionetworkv1alpha3.HTTPMatchRequest{
										{
											Headers: map[string]*istionetworkv1alpha3.StringMatch{
												"x-test-suite": {MatchType: &istionetworkv1alpha3.StringMatch_Exact{Exact: "feature-x"}},
											},
										},
										{
											Uri: &istionetworkv1alpha3.StringMatch{MatchType: &istionetworkv1alpha3.StringMatch_Exact{Exact: "/test-service"}},
										},
									},
									Rewrite: &istionetworkv1alpha3.HTTPRewrite{
										Uri: "/",
									},
								},
								{
									Route: []*istionetworkv1alpha3.HTTPRouteDestination{
										{
											Destination: &istionetworkv1alpha3.Destination{
												Host: "details",
											},
										},
									},
									Match: []*istionetworkv1alpha3.HTTPMatchRequest{
										{
											Uri: &istionetworkv1alpha3.StringMatch{MatchType: &istionetworkv1alpha3.StringMatch_Exact{Exact: "/test-service"}},
										},
									},
									Rewrite: &istionetworkv1alpha3.HTTPRewrite{
										Uri: "/",
									},
								},
								{
									Route: []*istionetworkv1alpha3.HTTPRouteDestination{
										{
											Destination: &istionetworkv1alpha3.Destination{
												Host:   "details",
												Subset: "v1",
											},
										},
									},
								},
							},
						},
					},
				}
			})

			Context("existing rule", func() {

				var (
					ref model.Ref
				)

				BeforeEach(func() {
					ref = model.Ref{
						Name:      "details-v1",
						Namespace: "test",
						Targets: []model.LocatedResourceStatus{
							model.NewLocatedResource("Deployment", "details-v1", map[string]string{"version": "v1"}),
							model.NewLocatedResource("Service", "details", map[string]string{}),
						},
						ResourceStatuses: []model.ResourceStatus{
							{Kind: VirtualServiceKind, Name: "details", Action: model.ActionModified},
						},
					}

				})

				It("should remove reference", func() {
					err := VirtualServiceRevertor(ctx, &ref)
					Expect(err).ToNot(HaveOccurred())

					virtualService := get.VirtualService("test", "details")
					Expect(reference.Get(&virtualService)).To(BeEmpty())
				})

				It("route removed", func() {
					err := VirtualServiceRevertor(ctx, &ref)
					Expect(err).ToNot(HaveOccurred())

					virtualService := get.VirtualService("test", "details")
					Expect(virtualService.Spec.Http).To(HaveLen(2))
				})

				It("correct route removed", func() {
					err := VirtualServiceRevertor(ctx, &ref)
					Expect(err).ToNot(HaveOccurred())

					virtualService := get.VirtualService("test", "details")
					Expect(virtualService.Spec.Http[0].Route[0].Destination.Subset).ToNot(Equal(model.GetSha("v1") + "-vs-test"))
				})
			})
		})

		Context("gateway attachment", func() {

			BeforeEach(func() {
				objects = []runtime.Object{
					&istionetwork.VirtualService{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "customer",
							Namespace: "test",
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
							Namespace: "test",
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
							Namespace: "test",
						},
						Spec: istionetworkv1alpha3.Gateway{
							Selector: map[string]string{"istio": "ingressgateway"},
							Servers: []*istionetworkv1alpha3.Server{
								{
									Port:  &istionetworkv1alpha3.Port{Name: "http", Protocol: "HTTP", Number: 80},
									Hosts: []string{"redhat-kubecon.io"},
								},
							},
						},
					},
				}
			})

			It("should attach to a host", func() {
				ref := model.Ref{
					Name: "customer-v1",
					Targets: []model.LocatedResourceStatus{
						model.NewLocatedResource("Service", "customer", nil),
						model.NewLocatedResource("Gateway", "test-gateway", map[string]string{LabelIkeHosts: "redhat-kubecon.io"}),
					},
				}

				err := VirtualServiceMutator(ctx, &ref)
				Expect(err).ToNot(HaveOccurred())

				created := get.VirtualService("test", "customer-"+ctx.Name)
				Expect(created.Spec.Hosts).To(ContainElement(ctx.Name + ".redhat-kubecon.io"))
			})

			It("should add request headers", func() {
				ref := model.Ref{
					Name: "customer-v1",
					Targets: []model.LocatedResourceStatus{
						model.NewLocatedResource("Service", "customer", nil),
						model.NewLocatedResource("Gateway", "test-gateway", map[string]string{LabelIkeHosts: "redhat-kubecon.io"}),
					},
				}

				err := VirtualServiceMutator(ctx, &ref)
				Expect(err).ToNot(HaveOccurred())

				created := get.VirtualService("test", "customer-"+ctx.Name)
				Expect(created.Spec.Http[0].Headers.Request.Add).To(HaveKeyWithValue(ctx.Route.Name, ctx.Route.Value))
			})

			It("should duplicate non effected vs", func() {
				ref := model.Ref{
					Name: "customer-v1",
					Targets: []model.LocatedResourceStatus{
						model.NewLocatedResource("Service", "customer", nil),
						model.NewLocatedResource("Gateway", "test-gateway", map[string]string{LabelIkeHosts: "redhat-kubecon.io"}),
					},
				}

				err := VirtualServiceMutator(ctx, &ref)
				Expect(err).ToNot(HaveOccurred())

				created := get.VirtualService("test", "non-customer-"+ctx.Name)
				Expect(created.Spec.Hosts).To(ContainElement(ctx.Name + ".redhat-kubecon.io"))
				Expect(created.Spec.Http[0].Headers.Request.Add).To(HaveKeyWithValue(ctx.Route.Name, ctx.Route.Value))
			})

		})
		Context("required", func() {
			var (
				virtualService istionetwork.VirtualService
			)

			BeforeEach(func() {
				virtualService = istionetwork.VirtualService{
					Spec: istionetworkv1alpha3.VirtualService{
						Http: []*istionetworkv1alpha3.HTTPRoute{
							{
								Route: []*istionetworkv1alpha3.HTTPRouteDestination{
									{
										Destination: &istionetworkv1alpha3.Destination{
											Host: "x",
										},
									},
									{
										Destination: &istionetworkv1alpha3.Destination{
											Host:   "y",
											Subset: "v1",
										},
									},
								},
							},
							{
								Route: []*istionetworkv1alpha3.HTTPRouteDestination{
									{
										Destination: &istionetworkv1alpha3.Destination{
											Host:   "z",
											Subset: "v2",
										},
									},
								},
							},
						},
					},
				}
			})
			It("Should require unversioned targets", func() {
				Expect(mutationRequired(virtualService, model.HostName{Name: "x"}, "v1")).To(BeTrue())
			})
			It("Should require versioned targets", func() {
				Expect(mutationRequired(virtualService, model.HostName{Name: "y"}, "v1")).To(BeTrue())
			})
			It("Should not require other versioned targets", func() {
				Expect(mutationRequired(virtualService, model.HostName{Name: "z"}, "v1")).To(BeFalse())
			})
		})
	})
})
