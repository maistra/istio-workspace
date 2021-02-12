package develop_test

import (
	istiov1alpha1 "github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/cmd/develop"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Hint", func() {

	var (
		kind, name, action = "Gateway", "test-gateway", "created"
		ref                = istiov1alpha1.RefStatus{
			Resources: []*istiov1alpha1.RefResource{
				{
					Kind:   &kind,
					Action: &action,
					Name:   &name,
					Prop: map[string]string{
						"hosts": "test.ike.io,test2.ike.io",
					},
				},
			},
		}
		route = istiov1alpha1.Route{
			Type:  "header",
			Name:  "x",
			Value: "y",
		}
	)

	It("should not print route if not route info provided", func() {
		text, err := develop.Hint(&ref, nil)
		Expect(err).ToNot(HaveOccurred())

		Expect(text).ToNot(ContainSubstring("the following header"))
	})

	It("should not print route if not header route info provided", func() {
		text, err := develop.Hint(nil, &istiov1alpha1.Route{Type: "nothing", Name: "x", Value: "y"})
		Expect(err).ToNot(HaveOccurred())

		Expect(text).ToNot(ContainSubstring("the following header"))
	})

	It("should print route if route provided", func() {
		text, err := develop.Hint(&ref, &route)
		Expect(err).ToNot(HaveOccurred())

		Expect(text).To(ContainSubstring("curl -H\"x:y\" YOUR_APP_URL."))
	})

	It("should print multiple hosts if hosts provided", func() {
		text, err := develop.Hint(&ref, &route)
		Expect(err).ToNot(HaveOccurred())

		Expect(text).To(ContainSubstring("curl test.ike.io"))
		Expect(text).To(ContainSubstring("curl test2.ike.io"))
	})

	It("should not print hosts if no hosts provided", func() {
		text, err := develop.Hint(nil, &route)
		Expect(err).ToNot(HaveOccurred())

		Expect(text).ToNot(ContainSubstring("the following hosts"))
	})
})
