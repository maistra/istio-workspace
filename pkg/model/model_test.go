package model_test

import (
	"github.com/maistra/istio-workspace/pkg/model"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
})
