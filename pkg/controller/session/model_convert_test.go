package session_test

import (
	"github.com/maistra/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/controller/session"
	"github.com/maistra/istio-workspace/pkg/model"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Basic model conversion", func() {

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

		Context("exists in status", func() {
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

			It("update status if existing found", func() {
				Expect(sess.Status.Refs).To(HaveLen(1))
				Expect(sess.Status.Refs[0].Resources).To(HaveLen(3))
				Expect(*sess.Status.Refs[0].Resources[0].Action).To(Equal("created"))
			})
		})
		Context("missing in status", func() {
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

			It("append to status if no name match", func() {
				Expect(sess.Status.Refs).To(HaveLen(2))
				Expect(sess.Status.Refs[0].Resources).To(HaveLen(1))
				Expect(*sess.Status.Refs[0].Resources[0].Action).To(Equal("failed"))
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
	Context("route to route", func() {
		var (
			route model.Route
		)
		JustBeforeEach(func() {
			route = session.RouteToRoute(&sess)
		})
		Context("missing", func() {
			BeforeEach(func() {
				sess = v1alpha1.Session{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-session",
					},
					Spec: v1alpha1.SessionSpec{
						Route: v1alpha1.Route{},
					},
					Status: v1alpha1.SessionStatus{},
				}
			})

			It("should default if no route defined", func() {
				Expect(route.Type).To(Equal(session.RouteStrategyHeader))
				Expect(route.Name).To(Equal(session.DefaultRouteHeaderName))
				Expect(route.Value).To(Equal(sess.ObjectMeta.Name))
			})
		})
		Context("exists", func() {
			BeforeEach(func() {
				sess = v1alpha1.Session{
					Spec: v1alpha1.SessionSpec{
						Route: v1alpha1.Route{
							Type:  "header",
							Name:  "x",
							Value: "y",
						},
					},
					Status: v1alpha1.SessionStatus{},
				}
			})

			It("should map route if provided", func() {
				Expect(route.Type).To(Equal(sess.Spec.Route.Type))
				Expect(route.Name).To(Equal(sess.Spec.Route.Name))
				Expect(route.Value).To(Equal(sess.Spec.Route.Value))
			})
		})

	})
})
