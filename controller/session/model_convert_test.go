package session_test

import (
	"github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/controller/session"
	"github.com/maistra/istio-workspace/pkg/model"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Basic model conversion", func() {

	var (
		kind, name, servname                   = "test", "1", "1-serv"
		aFailed, aModified, aCreated, aLocated = "failed", "modified", "created", "located"
		sess                                   v1alpha1.Session
	)
	Context("ref to status", func() {
		var (
			ref model.Ref
		)
		JustBeforeEach(func() {
			session.ConvertModelRefToAPIStatus(ref, &sess)

			Expect(sess.Status).ToNot(BeNil())
		})
		BeforeEach(func() {
			ref = model.Ref{
				Name:      "ref-name",
				Namespace: "x",
				Strategy:  "prepared-image",
				Args:      map[string]string{"image": "x"},
				Targets: []model.LocatedResourceStatus{
					{
						ResourceStatus: model.ResourceStatus{Kind: "dc", Name: "dc-n", Action: model.ActionLocated},
						Labels:         map[string]string{},
					},
					{
						ResourceStatus: model.ResourceStatus{Kind: "service", Name: "serv-n", Action: model.ActionLocated},
						Labels:         map[string]string{},
					}},
				ResourceStatuses: []model.ResourceStatus{
					{Kind: kind, Name: name, Action: model.ActionCreated},
					{Kind: "test", Name: "2", Action: model.ActionModified},
					{Kind: "test-2", Name: "2", Action: model.ActionFailed, Prop: map[string]string{"host": "x"}},
				}}
		})

		It("name mapped", func() {
			Expect(sess.Status.Refs).To(HaveLen(1))
			Expect(sess.Status.Refs[0].Name).To(Equal(ref.Name))
		})

		It("strategy mapped", func() {
			Expect(sess.Status.Refs).To(HaveLen(1))
			Expect(sess.Status.Refs[0].Strategy).To(Equal(ref.Strategy))
		})

		It("args mapped", func() {
			Expect(sess.Status.Refs).To(HaveLen(1))
			Expect(sess.Status.Refs[0].Args).To(Equal(ref.Args))
		})

		It("targets mapped", func() {
			Expect(sess.Status.Refs).To(HaveLen(1))
			Expect(sess.Status.Refs[0].Targets).To(HaveLen(2))

			Expect(sess.Status.Refs[0].Targets[0]).ToNot(BeNil())
			Expect(*sess.Status.Refs[0].Targets[0].Kind).To(Equal(ref.Targets[0].Kind))
			Expect(*sess.Status.Refs[0].Targets[0].Name).To(Equal(ref.Targets[0].Name))
			Expect(*sess.Status.Refs[0].Targets[0].Action).To(Equal("located"))

			Expect(sess.Status.Refs[0].Targets[1]).ToNot(BeNil())
			Expect(*sess.Status.Refs[0].Targets[1].Kind).To(Equal(ref.Targets[1].Kind))
			Expect(*sess.Status.Refs[0].Targets[1].Name).To(Equal(ref.Targets[1].Name))
			Expect(*sess.Status.Refs[0].Targets[1].Action).To(Equal("located"))
		})

		It("action mapped", func() {
			Expect(sess.Status.Refs).To(HaveLen(1))
			Expect(*sess.Status.Refs[0].Resources[0].Kind).To(Equal(ref.ResourceStatuses[0].Kind))
			Expect(*sess.Status.Refs[0].Resources[0].Name).To(Equal(ref.ResourceStatuses[0].Name))
			Expect(*sess.Status.Refs[0].Resources[0].Action).To(Equal(aCreated))
			Expect(sess.Status.Refs[0].Resources[0].Prop).To(BeEmpty())

			Expect(*sess.Status.Refs[0].Resources[1].Kind).To(Equal(ref.ResourceStatuses[1].Kind))
			Expect(*sess.Status.Refs[0].Resources[1].Name).To(Equal(ref.ResourceStatuses[1].Name))
			Expect(*sess.Status.Refs[0].Resources[1].Action).To(Equal(aModified))
			Expect(sess.Status.Refs[0].Resources[1].Prop).To(BeEmpty())

			Expect(*sess.Status.Refs[0].Resources[2].Kind).To(Equal(ref.ResourceStatuses[2].Kind))
			Expect(*sess.Status.Refs[0].Resources[2].Name).To(Equal(ref.ResourceStatuses[2].Name))
			Expect(*sess.Status.Refs[0].Resources[2].Action).To(Equal(aFailed))
			Expect(sess.Status.Refs[0].Resources[2].Prop).ToNot(BeEmpty())
			Expect(sess.Status.Refs[0].Resources[2].Prop["host"]).To(Equal("x"))
		})

		Context("exists in status", func() {
			BeforeEach(func() {
				sess = v1alpha1.Session{
					Status: v1alpha1.SessionStatus{
						Refs: []*v1alpha1.RefStatus{
							{
								Ref: v1alpha1.Ref{Name: ref.Name},
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
				Expect(*sess.Status.Refs[0].Resources[0].Action).To(Equal(aCreated))
			})
		})
		Context("missing in status", func() {
			BeforeEach(func() {
				sess = v1alpha1.Session{
					Status: v1alpha1.SessionStatus{
						Refs: []*v1alpha1.RefStatus{
							{
								Ref: v1alpha1.Ref{Name: ref.Name + "xxxx"},
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
				Expect(*sess.Status.Refs[0].Resources[0].Action).To(Equal(aFailed))
				Expect(sess.Status.Refs[1].Resources).To(HaveLen(3))
				Expect(*sess.Status.Refs[1].Resources[0].Action).To(Equal(aCreated))
			})
		})
	})

	Context("statuses to ref", func() {
		var (
			refs []*model.Ref
		)
		JustBeforeEach(func() {
			refs = session.ConvertAPIStatusesToModelRefs(sess)
		})
		BeforeEach(func() {
			sess = v1alpha1.Session{
				Status: v1alpha1.SessionStatus{
					Refs: []*v1alpha1.RefStatus{
						{
							Ref: v1alpha1.Ref{
								Name:     name + "xxxx",
								Strategy: "prepared-image",
								Args:     map[string]string{"image": "x"},
							},
							Targets: []*v1alpha1.LabeledRefResource{
								{
									RefResource: v1alpha1.RefResource{Kind: &kind, Name: &name, Action: &aLocated},
									Labels:      map[string]string{},
								},
								{
									RefResource: v1alpha1.RefResource{Kind: &kind, Name: &servname, Action: &aLocated},
									Labels:      map[string]string{},
								},
							},
							Resources: []*v1alpha1.RefResource{
								{
									Kind: &kind, Name: &name, Action: &aCreated,
								},
							},
						},
						{
							Ref: v1alpha1.Ref{Name: name + "xx"},
							Resources: []*v1alpha1.RefResource{
								{
									Kind: &kind, Name: &name, Action: &aFailed,
									Prop: map[string]string{"host": "x"},
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
			Expect(refs[0].Namespace).To(Equal(sess.Namespace))
			Expect(refs[0].Strategy).To(Equal(sess.Status.Refs[0].Strategy))
			Expect(refs[0].Args).To(Equal(sess.Status.Refs[0].Args))
			Expect(refs[0].ResourceStatuses[0].Kind).To(Equal(*sess.Status.Refs[0].Resources[0].Kind))
			Expect(refs[0].ResourceStatuses[0].Name).To(Equal(*sess.Status.Refs[0].Resources[0].Name))
			Expect(refs[0].ResourceStatuses[0].Action).To(Equal(model.ActionCreated))

			Expect(refs[1].Name).To(Equal(sess.Status.Refs[1].Name))
			Expect(refs[1].Namespace).To(Equal(sess.Namespace))
			Expect(refs[1].Strategy).To(Equal(sess.Status.Refs[1].Strategy))
			Expect(refs[1].Args).To(Equal(sess.Status.Refs[1].Args))
			Expect(refs[1].ResourceStatuses[0].Kind).To(Equal(*sess.Status.Refs[1].Resources[0].Kind))
			Expect(refs[1].ResourceStatuses[0].Name).To(Equal(*sess.Status.Refs[1].Resources[0].Name))
			Expect(refs[1].ResourceStatuses[0].Action).To(Equal(model.ActionFailed))
			Expect(refs[1].ResourceStatuses[0].Action).To(Equal(model.ActionFailed))
			Expect(refs[1].ResourceStatuses[0].Prop).ToNot(BeEmpty())
			Expect(refs[1].ResourceStatuses[0].Prop["host"]).To(Equal("x"))
		})

		It("targets mapped", func() {
			Expect(refs).To(HaveLen(2))

			Expect(refs[0].Name).To(Equal(sess.Status.Refs[0].Name))

			Expect(refs[0].Targets[0].Kind).To(Equal(*sess.Status.Refs[0].Targets[0].Kind))
			Expect(refs[0].Targets[0].Name).To(Equal(*sess.Status.Refs[0].Targets[0].Name))
			Expect(refs[0].Targets[0].Action).To(Equal(model.ActionLocated))

			Expect(refs[0].Targets[1].Kind).To(Equal(*sess.Status.Refs[0].Targets[1].Kind))
			Expect(refs[0].Targets[1].Name).To(Equal(*sess.Status.Refs[0].Targets[1].Name))
			Expect(refs[0].Targets[1].Action).To(Equal(model.ActionLocated))
		})
	})
	Context("route to route", func() {
		var (
			route model.Route
		)
		JustBeforeEach(func() {
			route = session.ConvertAPIRouteToModelRoute(&sess)
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
