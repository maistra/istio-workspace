package v1alpha1_test

import (
	"github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("API manipulation", func() {

	Context("when flipping readiness", func() {

		var components v1alpha1.StatusComponents
		componentOne := "one"
		componentTwo := "two"

		BeforeEach(func() {
			components = v1alpha1.StatusComponents{}
		})

		It("should set as pending", func() {
			components.SetPending(componentOne)

			Expect(components.Pending).To(HaveLen(1))
		})
		It("should set as Ready", func() {
			components.SetReady(componentOne)

			Expect(components.Ready).To(HaveLen(1))
		})
		It("should set as Unready", func() {
			components.SetUnready(componentOne)

			Expect(components.Unready).To(HaveLen(1))
		})

		It("should flip from pending to ready", func() {
			components.SetPending(componentOne)
			components.SetReady(componentOne)

			Expect(components.Ready).To(HaveLen(1))
			Expect(components.Pending).To(HaveLen(0))
		})
		It("should flip from pending to unready", func() {
			components.SetPending(componentOne)
			components.SetUnready(componentOne)

			Expect(components.Unready).To(HaveLen(1))
			Expect(components.Pending).To(HaveLen(0))
		})
		It("should flip from ready to unready", func() {
			components.SetReady(componentOne)
			components.SetUnready(componentOne)

			Expect(components.Unready).To(HaveLen(1))
			Expect(components.Ready).To(HaveLen(0))
		})

		It("should only flip one object", func() {
			components.SetReady(componentOne)
			components.SetReady(componentTwo)
			components.SetUnready(componentOne)

			Expect(components.Unready).To(HaveLen(1))
			Expect(components.Ready).To(HaveLen(1))
		})
	})
})
