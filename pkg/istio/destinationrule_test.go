package istio //nolint:testpackage //reason we want to test converters in isolation

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"

	"istio.io/api/networking/v1alpha3"
	k8yaml "sigs.k8s.io/yaml"
)

var _ = Describe("Operations for istio DestinationRule kind", func() {

	GetName := func(s *v1alpha3.Subset) string { return s.Name }

	var (
		err             error
		destinationRule istionetwork.DestinationRule
		yaml            string
	)

	Context("mutators", func() {

		JustBeforeEach(func() {
			err = k8yaml.Unmarshal([]byte(yaml), &destinationRule)
		})

		Context("existing rule", func() {
			var (
				mutatedDestinationRule istionetwork.DestinationRule
			)

			BeforeEach(func() {
				yaml = simpleDestinationRule
			})

			JustBeforeEach(func() {
				mutatedDestinationRule = mutateDestinationRule(destinationRule, "dr-test")
			})

			It("new subset added", func() {
				Expect(mutatedDestinationRule.Spec.Subsets).To(HaveLen(3))
			})

			It("new subset added with name", func() {
				Expect(mutatedDestinationRule.Spec.Subsets).To(ContainElement(WithTransform(GetName, Equal("dr-test"))))
			})

		})
	})

	Context("revertors", func() {

		JustBeforeEach(func() {
			err = k8yaml.Unmarshal([]byte(yaml), &destinationRule)
		})

		Context("existing rule", func() {
			var (
				revertedDestinationRule istionetwork.DestinationRule
			)

			BeforeEach(func() {
				yaml = simpleMutatedDestinationRule
			})

			It("new subset removed", func() {
				revertedDestinationRule = revertDestinationRule(destinationRule, "dr-test")
				Expect(err).ToNot(HaveOccurred())

				Expect(revertedDestinationRule.Spec.Subsets).To(HaveLen(2))
			})

			It("correct subset removed", func() {
				revertedDestinationRule = revertDestinationRule(destinationRule, "dr-test")
				Expect(err).ToNot(HaveOccurred())

				Expect(revertedDestinationRule.Spec.Subsets).ToNot(ContainElement(WithTransform(GetName, Equal("dr-test"))))
			})
		})
	})
})

var simpleDestinationRule = `kind: DestinationRule
metadata:
  annotations:
  creationTimestamp: 2019-01-16T19:28:05Z
  generation: 1
  name: details
  namespace: bookinfo
  resourceVersion: "4955188"
  selfLink: /apis/networking.istio.io/v1alpha3/namespaces/bookinfo/destinationrules/details
  uid: d928001c-19c4-11e9-a489-482ae3045b54
spec:
  host: details
  subsets:
  - labels:
      version: v1
    name: v1
  - labels:
      version: v2
    name: v2
`

var simpleMutatedDestinationRule = `kind: DestinationRule
metadata:
  creationTimestamp: "2019-01-16T19:28:05Z"
  generation: 1
  name: details
  namespace: bookinfo
  resourceVersion: "4955188"
  selfLink: /apis/networking.istio.io/v1alpha3/namespaces/bookinfo/destinationrules/details
  uid: d928001c-19c4-11e9-a489-482ae3045b54
spec:
  host: details
  subsets:
  - labels:
      version: v1
    name: v1
  - labels:
      version: v2
    name: v2
  - labels:
      version: dr-test
    name: dr-test
`
