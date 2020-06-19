package istio //nolint:testpackage //reason we want to test converters in isolation

import (
	"github.com/maistra/istio-workspace/pkg/model"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"

	k8yaml "sigs.k8s.io/yaml"
)

var _ = Describe("Operations for istio VirtualService kind", func() {

	Context("manipulation", func() {
		var (
			err            error
			virtualService istionetwork.VirtualService
			yaml           string
		)

		Context("mutators", func() {

			JustBeforeEach(func() {
				err = k8yaml.Unmarshal([]byte(yaml), &virtualService)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("existing rule", func() {
				GetMutatedRoute := func(vs istionetwork.VirtualService, host model.HostName, subset string) *v1alpha3.HTTPRoute {
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
					mutatedVirtualService istionetwork.VirtualService
					ctx                   model.SessionContext
					targetV1              = model.NewLocatedResource("Deployment", "details-v1", map[string]string{"version": "v1"})
					targetV1Host          = model.HostName{Name: "details"}
					targetV1Subset        = "v1-vs-test"
					targetV4              = model.NewLocatedResource("Deployment", "details-v4", map[string]string{"version": "v4"})
					targetV4Host          = model.HostName{Name: "details"}
					targetV4Subset        = "v4-vs-test"
					targetV5              = model.NewLocatedResource("Deployment", "details-v5", map[string]string{"version": "v5"})
					targetV5Host          = model.HostName{Name: "details"}
					targetV5Subset        = "v5-vs-test"
					targetV6              = model.NewLocatedResource("Deployment", "x-v5", map[string]string{"version": "v5"})
					targetV6Host          = model.HostName{Name: "x"}
					targetV6Subset        = "v5-vs-test"
				)

				BeforeEach(func() {
					yaml = complexVirtualService
					ctx = model.SessionContext{
						Name: "vs-test",
						Route: model.Route{
							Type:  "header",
							Name:  "test",
							Value: "x",
						},
					}
				})

				It("route added", func() {
					mutatedVirtualService, err = mutateVirtualService(ctx, targetV1Host, targetV1.Labels["version"], targetV1Subset, virtualService)
					Expect(err).ToNot(HaveOccurred())
					Expect(GetMutatedRoute(mutatedVirtualService, targetV1Host, targetV1Subset)).ToNot(BeNil())
				})

				It("route added before target route", func() {
					mutatedVirtualService, err = mutateVirtualService(ctx, targetV1Host, targetV1.Labels["version"], targetV1Subset, virtualService)
					Expect(err).ToNot(HaveOccurred())
					Expect(mutatedVirtualService.Spec.Http[0].Route[0].Destination.Subset).To(Equal(targetV1Subset))
				})

				It("has match", func() {
					mutatedVirtualService, err = mutateVirtualService(ctx, targetV1Host, targetV1.Labels["version"], targetV1Subset, virtualService)
					Expect(err).ToNot(HaveOccurred())
					Expect(GetMutatedRoute(mutatedVirtualService, targetV1Host, targetV1Subset).Match).ToNot(BeNil())
				})

				It("has subset", func() { // covered by GetMutatedRoute
					mutatedVirtualService, err = mutateVirtualService(ctx, targetV1Host, targetV1.Labels["version"], targetV1Subset, virtualService)
					Expect(err).ToNot(HaveOccurred())
					Expect(GetMutatedRoute(mutatedVirtualService, targetV1Host, targetV1Subset).Route[0].Destination.Subset).To(Equal(targetV1Subset))
				})

				It("create match when no match found", func() {
					mutatedVirtualService, err = mutateVirtualService(ctx, targetV4Host, targetV4.Labels["version"], targetV4Subset, virtualService)
					Expect(err).ToNot(HaveOccurred())

					mutated := GetMutatedRoute(mutatedVirtualService, targetV4Host, targetV4Subset)
					Expect(mutated).ToNot(BeNil())
					Expect(mutated.Match).To(HaveLen(1))
					for _, m := range mutated.Match {
						Expect(m.Headers).To(HaveLen(1))
						Expect(m.Headers["test"].GetExact()).To(Equal("x"))
					}
				})

				It("add route headers to found match with no headers", func() {
					mutatedVirtualService, err = mutateVirtualService(ctx, targetV1Host, targetV1.Labels["version"], targetV1Subset, virtualService)
					Expect(err).ToNot(HaveOccurred())

					mutated := GetMutatedRoute(mutatedVirtualService, targetV1Host, targetV1Subset)
					Expect(mutated).ToNot(BeNil())
					Expect(mutated.Match).To(HaveLen(2))
					for _, m := range mutated.Match {
						Expect(m.Headers).To(HaveLen(1))
						Expect(m.Headers["test"].GetExact()).To(Equal("x"))
					}
				})

				It("add route headers to found match with found headers", func() {
					mutatedVirtualService, err = mutateVirtualService(ctx, targetV5Host, targetV5.Labels["version"], targetV5Subset, virtualService)
					Expect(err).ToNot(HaveOccurred())

					mutated := GetMutatedRoute(mutatedVirtualService, targetV5Host, targetV5Subset)
					Expect(mutated).ToNot(BeNil())
					Expect(mutated.Match).To(HaveLen(1))
					for _, m := range mutated.Match {
						Expect(m.Headers).To(HaveLen(2))
						Expect(m.Headers["test"].GetExact()).To(Equal("x"))
					}
				})

				It("remove weighted destination", func() {
					mutatedVirtualService, err = mutateVirtualService(ctx, targetV1Host, targetV1.Labels["version"], targetV1Subset, virtualService)
					Expect(err).ToNot(HaveOccurred())

					Expect(GetMutatedRoute(mutatedVirtualService, targetV1Host, targetV1Subset).Route[0].Weight).To(Equal(int32(0)))
				})
				It("remove other destinations", func() {
					mutatedVirtualService, err = mutateVirtualService(ctx, targetV1Host, targetV1.Labels["version"], targetV1Subset, virtualService)
					Expect(err).ToNot(HaveOccurred())

					Expect(GetMutatedRoute(mutatedVirtualService, targetV1Host, targetV1Subset).Route).To(HaveLen(1))
				})
				It("remove mirror", func() {
					mutatedVirtualService, err = mutateVirtualService(ctx, targetV1Host, targetV1.Labels["version"], targetV1Subset, virtualService)
					Expect(err).ToNot(HaveOccurred())

					Expect(GetMutatedRoute(mutatedVirtualService, targetV1Host, targetV1Subset).Mirror).To(BeNil())
				})
				It("remove redirect", func() {
					mutatedVirtualService, err = mutateVirtualService(ctx, targetV1Host, targetV1.Labels["version"], targetV1Subset, virtualService)
					Expect(err).ToNot(HaveOccurred())

					Expect(GetMutatedRoute(mutatedVirtualService, targetV1Host, targetV1Subset).Redirect).To(BeNil())
				})
				It("include destinations with no subset", func() {
					mutatedVirtualService, err = mutateVirtualService(ctx, targetV6Host, targetV6.Labels["version"], targetV6Subset, virtualService)
					Expect(err).ToNot(HaveOccurred())

					Expect(GetMutatedRoute(mutatedVirtualService, targetV6Host, targetV6Subset).Redirect).To(BeNil())
				})

				It("route missing", func() {
					_, err = mutateVirtualService(ctx, model.HostName{Name: "miss-v5"}, "v5", "v5-vs-test", virtualService)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("route not found"))
				})
				It("route missing version", func() {
					_, err = mutateVirtualService(ctx, model.HostName{Name: "details-v10"}, "v10", "v10-vs-test", virtualService)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("route not found"))
				})
			})
		})

		Context("revertors", func() {

			JustBeforeEach(func() {
				err = k8yaml.Unmarshal([]byte(yaml), &virtualService)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("existing rule", func() {
				var revertedVirtualService istionetwork.VirtualService

				BeforeEach(func() {
					yaml = complextMutatedVirtualService
				})

				JustBeforeEach(func() {
					revertedVirtualService = revertVirtualService("v1-vs-test", virtualService)
				})

				It("route removed", func() {
					Expect(revertedVirtualService.Spec.Http).To(HaveLen(2))
				})

				It("correct route removed", func() {
					Expect(revertedVirtualService.Spec.Http[0].Route[0].Destination.Subset).ToNot(Equal("v1-vs-test"))
				})
			})
		})
	})

	Context("required", func() {
		var (
			virtualService istionetwork.VirtualService
		)

		BeforeEach(func() {
			virtualService = istionetwork.VirtualService{
				Spec: v1alpha3.VirtualService{
					Http: []*v1alpha3.HTTPRoute{
						{
							Route: []*v1alpha3.HTTPRouteDestination{
								{
									Destination: &v1alpha3.Destination{
										Host: "x",
									},
								},
								{
									Destination: &v1alpha3.Destination{
										Host:   "y",
										Subset: "v1",
									},
								},
							},
						},
						{
							Route: []*v1alpha3.HTTPRouteDestination{
								{
									Destination: &v1alpha3.Destination{
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

var complexVirtualService = `kind: VirtualService
metadata:
  name: details
  namespace: bookinfo
spec:
  hosts:
  - details
  http:
  - route:
    - destination:
        host: details
        subset: v1
      weight: 50
    - destination:
        host: details
        subset: v2
      weight: 50
    match:
      - uri:
          prefix: /a
      - uri:
          prefix: /b
    mirror:
      host: details
      subset: v3
    redirect:
      uri: /redirected
  - route:
    - destination:
        host: details
        subset: v4
  - route:
    - destination:
        host: details
        subset: v5
    match:
      - headers:
          request-id:
            exact: test
  - route:
    - destination:
        host: x
`

var complextMutatedVirtualService = `apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  creationTimestamp: "2020-01-15T22:53:54Z"
  generation: 11
  name: reviews
  namespace: aslak-devconf
  resourceVersion: "591982"
  selfLink: /apis/networking.istio.io/v1alpha3/namespaces/aslak-devconf/virtualservices/reviews
  uid: e7a2a377-37e9-11ea-bd3f-02fb5d7d8a95
spec:
  hosts:
  - details
  http:
  - match:
    - headers:
        x-test-suite:
          exact: feature-x
    route:
    - destination:
        host: details
        subset: v1-vs-test
  - match:
    - headers:
        x-test-suite:
          exact: feature-x
      uri:
        prefix: /test-service
    rewrite:
      uri: /
    route:
    - destination:
        host: details
        subset: v1-vs-test
  - match:
    - uri:
        prefix: /test-service
    rewrite:
      uri: /
    route:
    - destination:
        host: details
  - route:
    - destination:
        host: details
        subset: v1`
