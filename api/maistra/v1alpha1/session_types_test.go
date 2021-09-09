package v1alpha1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/maistra/istio-workspace/api/maistra/v1alpha1"
)

var _ = Describe("API manipulation", func() {

	Context("when flipping readiness", func() {

		var components v1alpha1.StatusComponents
		compoentOne := "one"
		compoentTwo := "two"

		BeforeEach(func() {
			components = v1alpha1.StatusComponents{}
		})

		It("should set as pending", func() {
			components.SetPending(compoentOne)

			Expect(components.Pending).To(HaveLen(1))
		})
		It("should set as Ready", func() {
			components.SetReady(compoentOne)

			Expect(components.Ready).To(HaveLen(1))
		})
		It("should set as UnReady", func() {
			components.SetUnReady(compoentOne)

			Expect(components.UnReady).To(HaveLen(1))
		})

		It("should flip from pending to ready", func() {
			components.SetPending(compoentOne)
			Expect(components.Pending).To(HaveLen(1))

			components.SetReady(compoentOne)
			Expect(components.Ready).To(HaveLen(1))
			Expect(components.Pending).To(HaveLen(0))
		})
		It("should flip from pending to unready", func() {
			components.SetPending(compoentOne)
			Expect(components.Pending).To(HaveLen(1))

			components.SetUnReady(compoentOne)
			Expect(components.UnReady).To(HaveLen(1))
			Expect(components.Pending).To(HaveLen(0))
		})
		It("should flip from ready to unready", func() {
			components.SetReady(compoentOne)
			Expect(components.Ready).To(HaveLen(1))

			components.SetUnReady(compoentOne)
			Expect(components.UnReady).To(HaveLen(1))
			Expect(components.Ready).To(HaveLen(0))
		})

		It("should only flip one object", func() {
			components.SetReady(compoentOne)
			components.SetReady(compoentTwo)
			Expect(components.Ready).To(HaveLen(2))

			components.SetUnReady(compoentOne)
			Expect(components.UnReady).To(HaveLen(1))
			Expect(components.Ready).To(HaveLen(1))
		})
	})
})
