package model_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/maistra/istio-workspace/pkg/model"
)

var _ = Describe("Operations for model object", func() {

	Context("of type hostname", func() {

		It("should match on short name", func() {
			h := model.HostName{Name: "x", Namespace: "y"}
			Expect(h.Match("x")).To(BeTrue())
		})
		It("should match on full name in same namespace", func() {
			h := model.HostName{Name: "x", Namespace: "y"}
			Expect(h.Match("x.y.svc.cluster.local")).To(BeTrue())
		})
		It("should not match on different short name", func() {
			h := model.HostName{Name: "x", Namespace: "y"}
			Expect(h.Match("y")).To(BeFalse())
		})
		It("should not match on full name in different namespace", func() {
			h := model.HostName{Name: "x", Namespace: "y"}
			Expect(h.Match("x.z.svc.cluster.local")).To(BeFalse())
		})

	})

	Context("refkindname parsing", func() {

		It("should parse name", func() {
			ref := model.ParseRefKindName("name")
			Expect(ref.Name).To(Equal("name"))
			Expect(ref.Kind).To(BeEmpty())
		})

		It("should parse and trim name", func() {
			ref := model.ParseRefKindName("      					name		    					")
			Expect(ref.Name).To(Equal("name"))
			Expect(ref.Kind).To(BeEmpty())
		})

		It("should parse kind and name removing spacing characters", func() {
			ref := model.ParseRefKindName(" dc/name123    ")
			Expect(ref.Name).To(Equal("name123"))
			Expect(ref.Kind).To(Equal("dc"))
		})

		It("should parse name only when more than one / present in the expression", func() {
			ref := model.ParseRefKindName("dc/marvel/name123")
			Expect(ref.Name).To(Equal("dc/marvel/name123"))
			Expect(ref.Kind).To(BeEmpty())
		})

	})
})
