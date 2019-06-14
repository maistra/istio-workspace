package istio

import (
	"github.com/maistra/istio-workspace/pkg/model"
	k8yaml "sigs.k8s.io/yaml"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/api/pkg/kube/apis/networking/v1alpha3"
)

var _ = Describe("Operations for istio VirtualService kind", func() {

	/*
		Context("discovery", func() {
			var (
				objects []runtime.Object
				ctx     model.SessionContext
				ref     *model.Ref
				err     error
			)
			JustBeforeEach(func() {
				ctx = model.SessionContext{
					Context:   context.TODO(),
					Name:      "test",
					Namespace: "test",
					Log:       logf.Log.WithName("test"),
					Client:    fake.NewFakeClient(objects...),
				}
			})

			Context("mutators", func() {
				BeforeEach(func() {
					ref = &model.Ref{Name: "test"}
					objects = []runtime.Object{
						&istionetwork.VirtualService{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test",
								Namespace: "test",
							},
							Spec: v1alpha3.VirtualService{
								Hosts: []string{},
								Http: []*v1alpha3.HTTPRoute{
									&v1alpha3.HTTPRoute{
										Route: []*v1alpha3.HTTPRouteDestination{
											&v1alpha3.HTTPRouteDestination{
												Destination: &v1alpha3.Destination{
													Host: "test",
												},
												Weight: 10,
											},
										},
									},
								},
							},
						},
					}
				})
			})
		})
	*/
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
				GetMutatedRoute := func(vs istionetwork.VirtualService, host, subset string) *v1alpha3.HTTPRoute {
					for _, h := range vs.Spec.Http {
						for _, r := range h.Route {
							if r.Destination.Host == host && r.Destination.Subset == subset {
								return h
							}
						}
					}
					return nil
				}
				var (
					mutatedVirtualService istionetwork.VirtualService
					ctx                   model.SessionContext
					targetHost            = "details"
					targetSubset          = "v1-vs-test"
					locatedTarget         = model.NewLocatedResource("Deployment", "details-v1", map[string]string{"version": "v1"})
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

				JustBeforeEach(func() {
					mutatedVirtualService, err = mutateVirtualService(ctx, locatedTarget, virtualService)
					Expect(err).ToNot(HaveOccurred())
				})
				It("route added", func() {
					Expect(GetMutatedRoute(mutatedVirtualService, targetHost, targetSubset)).ToNot(BeNil())
				})

				It("has match", func() {
					Expect(GetMutatedRoute(mutatedVirtualService, targetHost, targetSubset).Match).ToNot(BeNil())
				})

				It("has subset", func() { // covered by GetMutatedRoute
					Expect(GetMutatedRoute(mutatedVirtualService, targetHost, targetSubset).Route[0].Destination.Subset).To(Equal(targetSubset))
				})

				It("add headers to matches", func() {
					mutated := GetMutatedRoute(mutatedVirtualService, targetHost, targetSubset)
					Expect(mutated).ToNot(BeNil())
					Expect(mutated.Match).To(HaveLen(2))
					for _, m := range mutated.Match {
						Expect(m.Headers).To(HaveLen(1)) // also validate we can keep other existing headers
						Expect(m.Headers["test"].GetExact()).To(Equal("x"))
					}
				})

				It("remove weighted destination", func() {
					Expect(GetMutatedRoute(mutatedVirtualService, targetHost, targetSubset).Route[0].Weight).To(Equal(int32(0)))
				})
				It("remove other destinations", func() {
					Expect(GetMutatedRoute(mutatedVirtualService, targetHost, targetSubset).Route).To(HaveLen(1))
				})
				It("remove mirror", func() {
					Expect(GetMutatedRoute(mutatedVirtualService, targetHost, targetSubset).Mirror).To(BeNil())
				})
				It("remove redirect", func() {
					Expect(GetMutatedRoute(mutatedVirtualService, targetHost, targetSubset).Redirect).To(BeNil())
				})
			})
		})

		Context("revertors", func() {

			JustBeforeEach(func() {
				err = k8yaml.Unmarshal([]byte(yaml), &virtualService)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("existing rule", func() {
				var (
					revertedVirtualService istionetwork.VirtualService
					ctx                    model.SessionContext
				)

				BeforeEach(func() {
					yaml = simpleMutatedVirtualService
					ctx = model.SessionContext{
						Route: model.Route{
							Type:  "header",
							Name:  "test",
							Value: "x",
						},
					}
				})

				JustBeforeEach(func() {
					revertedVirtualService, err = revertVirtualService(ctx, "v1-vs-test", virtualService)
					Expect(err).ToNot(HaveOccurred())
				})

				It("route removed", func() {
					Expect(revertedVirtualService.Spec.Http).To(HaveLen(1))
				})

				It("correct route removed", func() {
					Expect(revertedVirtualService.Spec.Http[0].Route[0].Destination.Subset).ToNot(Equal("v1-vs-test"))
				})
			})
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
        host: x
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
        subset: v1-vs-test
  - route:
    - destination:
        host: details
        subset: v1
`
