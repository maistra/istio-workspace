package model_test

import (
	"fmt"

	"github.com/maistra/istio-workspace/pkg/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Operations for model object", func() {

	Context("FIXME: sort and hashing", func() {

		It("locator store sort", func() {

			store := model.LocatorStore{}
			store.Report(model.LocatorStatus{Resource: model.Resource{Name: "X", Kind: "X"}, Action: model.ActionCreate})
			store.Report(model.LocatorStatus{Resource: model.Resource{Name: "X", Kind: "X"}, Action: model.ActionDelete})
			store.Report(model.LocatorStatus{Resource: model.Resource{Name: "Y", Kind: "X"}, Action: model.ActionDelete})
			store.Report(model.LocatorStatus{Resource: model.Resource{Name: "Y", Kind: "X"}, Action: model.ActionCreate})

			ls := store.Store("X")
			for i, l := range ls {
				if (i == 0 || i == 1) && l.Action != model.ActionDelete {
					GinkgoT().Error("should sort Delete first")
				}
				if (i == 2 || i == 3) && l.Action != model.ActionCreate {
					GinkgoT().Error("should sort Delete first")
				}
				fmt.Println(l)
			}
		})

		It("ref hash", func() {

			r1 := model.Ref{KindName: model.ParseRefKindName("x"), Namespace: "Y", Strategy: "X", Remove: true, Args: map[string]string{"X": "Y", "Y": "X"}}
			r2 := model.Ref{KindName: model.ParseRefKindName("x"), Namespace: "Y", Strategy: "X", Remove: true, Args: map[string]string{"Y": "X", "X": "Y"}}

			refs := []model.Ref{
				{KindName: model.ParseRefKindName("x"), Namespace: "Y", Strategy: "X", Remove: true, Args: map[string]string{"Y": "X", "X": "X"}},
				{KindName: model.ParseRefKindName("x"), Namespace: "Y", Strategy: "X", Remove: false, Args: map[string]string{"Y": "X", "X": "Y"}},
				{KindName: model.ParseRefKindName("x"), Namespace: "Y", Strategy: "Y", Remove: true, Args: map[string]string{"Y": "X", "X": "Y"}},
				{KindName: model.ParseRefKindName("x"), Namespace: "X", Strategy: "X", Remove: true, Args: map[string]string{"Y": "X", "X": "Y"}},
				{KindName: model.ParseRefKindName("Y"), Namespace: "Y", Strategy: "X", Remove: true, Args: map[string]string{"Y": "X", "X": "Y"}},
				{KindName: model.ParseRefKindName("Y"), Namespace: "Y", Strategy: "X", Remove: true},
				{KindName: model.ParseRefKindName("Y"), Namespace: "Y", Strategy: "X"},
				{KindName: model.ParseRefKindName("Y"), Namespace: "Y"},
				{KindName: model.ParseRefKindName("Y")},
			}

			if r1.Hash() != r2.Hash() {
				GinkgoT().Errorf("Should match %v %v %v", r1.Hash() == r2.Hash(), r1.Hash(), r2.Hash())
			}

			for _, r := range refs {
				if r1.Hash() == r.Hash() {
					GinkgoT().Errorf("Should match %v %v %v", r1.Hash() == r.Hash(), r1.Hash(), r.Hash())
				}
			}
		})

	})

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
