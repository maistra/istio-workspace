package istio

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	istionetwork "istio.io/api/pkg/kube/apis/networking/v1alpha3"

	k8yaml "sigs.k8s.io/yaml"
)

var _ = Describe("Operations for istio VirtualService kind", func() {

	var (
		err            error
		virtualService istionetwork.VirtualService
		yaml           string
	)

	Context("mutators", func() {

		JustBeforeEach(func() {
			err = k8yaml.Unmarshal([]byte(yaml), &virtualService)
		})

		Context("existing rule", func() {
			var (
				mutatedVirtualService istionetwork.VirtualService
			)

			BeforeEach(func() {
				yaml = simpleVirtualService
			})

			JustBeforeEach(func() {
				mutatedVirtualService, err = mutateVirtualService(virtualService)
				Expect(err).ToNot(HaveOccurred())
			})

			It("route added", func() {
				Expect(mutatedVirtualService.Spec.Http).To(HaveLen(2))
			})
			It("first route has match", func() {
				Expect(mutatedVirtualService.Spec.Http[0].Match).ToNot(BeNil())
			})
			It("first route has subset", func() {
				Expect(mutatedVirtualService.Spec.Http[0].Route[0].Destination.Subset).To(Equal("v1-test"))
			})
		})
	})

	Context("revertors", func() {

		JustBeforeEach(func() {
			err = k8yaml.Unmarshal([]byte(yaml), &virtualService)
		})

		Context("existing rule", func() {
			var (
				revertedVirtualService istionetwork.VirtualService
			)

			BeforeEach(func() {
				yaml = simpleMutatedVirtualService
			})

			JustBeforeEach(func() {
				revertedVirtualService, err = revertVirtualService(virtualService)
				Expect(err).ToNot(HaveOccurred())
			})

			It("route removed", func() {
				Expect(revertedVirtualService.Spec.Http).To(HaveLen(1))
			})
			It("correct route removed", func() {
				Expect(revertedVirtualService.Spec.Http[0].Route[0].Destination.Subset).ToNot(Equal("v1-test"))
			})
		})
	})
})

var simpleVirtualService = `kind: VirtualService
metadata:
  annotations:
  creationTimestamp: 2019-01-16T20:58:51Z
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
  - route:
    - destination:
        host: details
        subset: v1
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
        subset: v1-test
  - route:
    - destination:
        host: details
        subset: v1
`
