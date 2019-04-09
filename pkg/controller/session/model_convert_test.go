package session_test

import (
	"github.com/aslakknutsen/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/aslakknutsen/istio-workspace/pkg/controller/session"
	"github.com/aslakknutsen/istio-workspace/pkg/model"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Basic model convertion", func() {

	var (
		kind, name, aFailed, aCreated = "test", "1", "failed", "created"
		sess                          v1alpha1.Session
	)
	Context("ref to status", func() {
		var (
			ref model.Ref
		)
		JustBeforeEach(func() {
			session.RefToStatus(ref, &sess)

			Expect(sess.Status).ToNot(BeNil())
		})
		BeforeEach(func() {
			ref = model.Ref{
				Name: "ref-name",
				ResourceStatuses: []model.ResourceStatus{
					{Kind: "test", Name: "1", Action: model.ActionCreated},
					{Kind: "test", Name: "2", Action: model.ActionModified},
					{Kind: "test-2", Name: "2", Action: model.ActionFailed},
				}}
		})

		It("name mapped", func() {
			Expect(sess.Status.Refs).To(HaveLen(1))
			Expect(sess.Status.Refs[0].Name).To(Equal(ref.Name))
		})

		It("action mapped", func() {
			Expect(sess.Status.Refs).To(HaveLen(1))
			Expect(*sess.Status.Refs[0].Resources[0].Kind).To(Equal(ref.ResourceStatuses[0].Kind))
			Expect(*sess.Status.Refs[0].Resources[0].Name).To(Equal(ref.ResourceStatuses[0].Name))
			Expect(*sess.Status.Refs[0].Resources[0].Action).To(Equal("created"))

			Expect(*sess.Status.Refs[0].Resources[1].Kind).To(Equal(ref.ResourceStatuses[1].Kind))
			Expect(*sess.Status.Refs[0].Resources[1].Name).To(Equal(ref.ResourceStatuses[1].Name))
			Expect(*sess.Status.Refs[0].Resources[1].Action).To(Equal("modified"))

			Expect(*sess.Status.Refs[0].Resources[2].Kind).To(Equal(ref.ResourceStatuses[2].Kind))
			Expect(*sess.Status.Refs[0].Resources[2].Name).To(Equal(ref.ResourceStatuses[2].Name))
			Expect(*sess.Status.Refs[0].Resources[2].Action).To(Equal("failed"))
		})

		Context("ref update based on name", func() {
			BeforeEach(func() {
				sess = v1alpha1.Session{
					Status: v1alpha1.SessionStatus{
						Refs: []*v1alpha1.RefStatus{
							{
								Name: ref.Name,
								Resources: []*v1alpha1.RefResource{
									{
										Kind: &kind, Name: &name, Action: &aFailed,
									},
								},
							},
						},
					},
				}
			})

			It("update ref in status if changed", func() {
				Expect(sess.Status.Refs).To(HaveLen(1))
				Expect(sess.Status.Refs[0].Resources).To(HaveLen(3))
				Expect(*sess.Status.Refs[0].Resources[0].Action).To(Equal("created"))
			})
		})
		Context("ref add when different name", func() {
			BeforeEach(func() {
				sess = v1alpha1.Session{
					Status: v1alpha1.SessionStatus{
						Refs: []*v1alpha1.RefStatus{
							{
								Name: ref.Name + "xxxx",
								Resources: []*v1alpha1.RefResource{
									{
										Kind: &kind, Name: &name, Action: &aFailed,
									},
								},
							},
						},
					},
				}
			})

			It("append ref in status if no name match", func() {
				Expect(sess.Status.Refs).To(HaveLen(2))
				Expect(sess.Status.Refs[0].Resources).To(HaveLen(1))
				Expect(sess.Status.Refs[1].Resources).To(HaveLen(3))
				Expect(*sess.Status.Refs[1].Resources[0].Action).To(Equal("created"))
			})
		})
	})

	Context("statuses to ref", func() {
		var (
			refs []*model.Ref
		)
		JustBeforeEach(func() {
			refs = session.StatusesToRef(sess)
		})
		BeforeEach(func() {
			sess = v1alpha1.Session{
				Status: v1alpha1.SessionStatus{
					Refs: []*v1alpha1.RefStatus{
						{
							Name: name,
							Resources: []*v1alpha1.RefResource{
								{
									Kind: &kind, Name: &name, Action: &aCreated,
								},
							},
						},
						{
							Name: name + "xx",
							Resources: []*v1alpha1.RefResource{
								{
									Kind: &kind, Name: &name, Action: &aFailed,
								},
							},
						},
					},
				},
			}
		})

		It("convert all refs", func() {
			Expect(refs).To(HaveLen(2))

			Expect(refs[0].Name).To(Equal(sess.Status.Refs[0].Name))
			Expect(refs[0].ResourceStatuses[0].Kind).To(Equal(*sess.Status.Refs[0].Resources[0].Kind))
			Expect(refs[0].ResourceStatuses[0].Name).To(Equal(*sess.Status.Refs[0].Resources[0].Name))
			Expect(refs[0].ResourceStatuses[0].Action).To(Equal(model.ActionCreated))

			Expect(refs[1].Name).To(Equal(sess.Status.Refs[1].Name))
			Expect(refs[1].ResourceStatuses[0].Kind).To(Equal(*sess.Status.Refs[1].Resources[0].Kind))
			Expect(refs[1].ResourceStatuses[0].Name).To(Equal(*sess.Status.Refs[1].Resources[0].Name))
			Expect(refs[1].ResourceStatuses[0].Action).To(Equal(model.ActionFailed))
		})

	})
})
