package develop_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	istiov1alpha1 "github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/cmd/develop"
	"github.com/maistra/istio-workspace/pkg/session"
)

var _ = Describe("Hint", func() {

	var (
		hosts = []string{"test.ike.io", "test2.ike.io"}
		route = istiov1alpha1.Route{
			Type:  "header",
			Name:  "x",
			Value: "y",
		}
		validState = session.State{
			Hosts: hosts,
			Route: route,
		}
	)

	It("should not print route if not route info provided", func() {
		text, err := develop.Hint(&session.State{
			Hosts: hosts,
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(text).ToNot(ContainSubstring("the following header"))
	})

	It("should not print route if not header route info provided", func() {
		text, err := develop.Hint(&session.State{
			Route: istiov1alpha1.Route{Type: "nothing", Name: "x", Value: "y"},
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(text).ToNot(ContainSubstring("the following header"))
	})

	It("should print route if route provided", func() {
		text, err := develop.Hint(&validState)
		Expect(err).ToNot(HaveOccurred())

		Expect(text).To(ContainSubstring("curl -H\"x:y\" YOUR_APP_URL."))
	})

	It("should print multiple hosts if hosts provided", func() {
		text, err := develop.Hint(&validState)
		Expect(err).ToNot(HaveOccurred())

		Expect(text).To(ContainSubstring("curl test.ike.io"))
		Expect(text).To(ContainSubstring("curl test2.ike.io"))
	})

	It("should not print hosts if no hosts provided", func() {
		text, err := develop.Hint(&session.State{
			Route: route,
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(text).ToNot(ContainSubstring("the following hosts"))
	})
})
