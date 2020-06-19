package session_test

import (
	"context"
	"time"

	"github.com/maistra/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/controller/session"
	"github.com/maistra/istio-workspace/pkg/model"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	kind, name, action = "test", "details", "created"
)

var _ = Describe("Basic session manipulation", func() {
	var (
		objects    []runtime.Object
		controller reconcile.Reconciler
		req        reconcile.Request
		schema     *runtime.Scheme
		c          client.Client
		locator    *trackedLocator
		mutator    *trackedMutator
		revertor   *trackedRevertor
	)
	GetSession := func(c *client.Client) func(namespace, name string) v1alpha1.Session {
		return func(namespace, name string) v1alpha1.Session {
			s := v1alpha1.Session{}
			err := (*c).Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &s)
			Expect(err).ToNot(HaveOccurred())
			return s
		}
	}(&c)
	GetStatusRef := func(name string, session v1alpha1.Session) *v1alpha1.RefStatus {
		for _, ref := range session.Status.Refs {
			if ref.Name == name {
				return ref
			}
		}
		return nil
	}

	JustBeforeEach(func() {
		locator = &trackedLocator{Action: notFoundTestLocator}
		mutator = &trackedMutator{Action: noOp}
		revertor = &trackedRevertor{Action: noOp}
		manipulators := session.Manipulators{
			Locators:  []model.Locator{locator.Do},
			Mutators:  []model.Mutator{mutator.Do},
			Revertors: []model.Revertor{revertor.Do},
		}

		schema, _ = v1alpha1.SchemeBuilder.Build()
		req = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      "test-session",
				Namespace: "test",
			},
		}
		c = fake.NewFakeClientWithScheme(schema, objects...)
		controller = session.NewStandaloneReconciler(c, manipulators)
	})

	Context("session creation", func() {

		Context("session not found", func() {
			BeforeEach(func() {
				objects = []runtime.Object{}
			})

			It("no retry", func() {
				res, err := controller.Reconcile(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())
			})
		})
		Context("session found", func() {
			BeforeEach(func() {
				objects = []runtime.Object{
					&v1alpha1.Session{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-session",
							Namespace: "test",
						},
						Spec: v1alpha1.SessionSpec{
							Refs: []v1alpha1.Ref{{Name: "details"}},
						},
					},
				}
			})

			It("finalizer added", func() {
				res, err := controller.Reconcile(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				modified := GetSession("test", "test-session")
				Expect(modified.ObjectMeta.Finalizers).To(HaveLen(1))
			})

			It("mutate when target is located", func() {
				locator.Action = foundTestLocator

				res, err := controller.Reconcile(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				Expect(locator.WasCalled).To(BeTrue())
				Expect(mutator.WasCalled).To(BeTrue())
			})
			It("revertors not called when mutation occure", func() {
				locator.Action = foundTestLocator
				mutator.Action = addResourceStatus(model.ResourceStatus{Name: "details", Kind: "test", Action: model.ActionCreated})

				res, err := controller.Reconcile(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				Expect(locator.WasCalled).To(BeTrue())
				Expect(mutator.WasCalled).To(BeTrue())
				Expect(revertor.WasCalled).ToNot(BeTrue())
			})
			It("status is updated when mutation occure", func() {
				locator.Action = foundTestLocator
				mutator.Action = addResourceStatus(model.ResourceStatus{Name: "details", Kind: "test", Action: model.ActionCreated})

				res, err := controller.Reconcile(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				modified := GetSession("test", "test-session")
				Expect(modified.Status).ToNot(BeNil())
				Expect(modified.Status.Refs).To(HaveLen(1))
				Expect(modified.Status.Refs[0].Name).To(Equal("details"))
			})
		})
	})

	Context("session modification", func() {
		Context("new reference", func() {
			BeforeEach(func() {
				objects = []runtime.Object{
					&v1alpha1.Session{
						ObjectMeta: metav1.ObjectMeta{
							Name:       "test-session",
							Namespace:  "test",
							Finalizers: []string{session.Finalizer},
						},
						Spec: v1alpha1.SessionSpec{
							Refs: []v1alpha1.Ref{{Name: "details"}, {Name: "details2"}},
						},
						Status: v1alpha1.SessionStatus{
							Refs: []*v1alpha1.RefStatus{
								{
									Ref:       v1alpha1.Ref{Name: "details"},
									Resources: []*v1alpha1.RefResource{{Kind: &kind, Name: &name, Action: &action}},
								},
							},
						},
					},
				}
			})
			It("revertors not called when mutation occure", func() {
				locator.Action = foundTestLocator
				mutator.Action = addResourceStatus(model.ResourceStatus{Name: "details2", Kind: "test", Action: model.ActionCreated})

				res, err := controller.Reconcile(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				Expect(locator.WasCalled).To(BeTrue())
				Expect(mutator.WasCalled).To(BeTrue())
				Expect(revertor.WasCalled).ToNot(BeTrue())
			})
			It("existing status is updated when new mutation occure", func() {
				locator.Action = foundTestLocator
				mutator.Action = addResourceStatus(model.ResourceStatus{Name: "details2", Kind: "test", Action: model.ActionCreated})

				res, err := controller.Reconcile(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				modified := GetSession("test", "test-session")
				Expect(modified.Status).ToNot(BeNil())
				Expect(modified.Status.Refs).To(HaveLen(2))
				Expect(modified.Status.Refs[0].Name).To(Equal("details"))
				Expect(modified.Status.Refs[1].Name).To(Equal("details2"))
			})
		})
		Context("removed reference", func() {
			BeforeEach(func() {
				objects = []runtime.Object{
					&v1alpha1.Session{
						ObjectMeta: metav1.ObjectMeta{
							Name:       "test-session",
							Namespace:  "test",
							Finalizers: []string{session.Finalizer},
						},
						Spec: v1alpha1.SessionSpec{
							Refs: []v1alpha1.Ref{},
						},
						Status: v1alpha1.SessionStatus{
							Refs: []*v1alpha1.RefStatus{
								{
									Ref:       v1alpha1.Ref{Name: "details"},
									Resources: []*v1alpha1.RefResource{{Kind: &kind, Name: &name, Action: &action}},
								},
							},
						},
					},
				}
			})
			It("revertors called when ref removed", func() {
				locator.Action = foundTestLocator
				revertor.Action = removeResourceStatus("test", "details")
				res, err := controller.Reconcile(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				Expect(locator.WasCalled).To(BeFalse())
				Expect(mutator.WasCalled).To(BeFalse())
				Expect(revertor.WasCalled).To(BeTrue())
			})

			It("status removed when ref removed", func() {
				locator.Action = foundTestLocator
				revertor.Action = removeResourceStatus("test", "details")
				res, err := controller.Reconcile(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				modified := GetSession("test", "test-session")
				Expect(modified.Status).ToNot(BeNil())
				Expect(modified.Status.Refs).To(HaveLen(0))
			})
		})
		Context("updated reference", func() {
			locatedTargetName := "test-a"
			locatedAction := "located"
			BeforeEach(func() {
				objects = []runtime.Object{
					&v1alpha1.Session{
						ObjectMeta: metav1.ObjectMeta{
							Name:       "test-session",
							Namespace:  "test",
							Finalizers: []string{session.Finalizer},
						},
						Spec: v1alpha1.SessionSpec{
							Refs: []v1alpha1.Ref{
								{
									Name:     "details",
									Strategy: "telepresence",
								},
								{
									Name:     "ratings",
									Strategy: "prepared-image",
									Args: map[string]string{
										"image": "x",
									},
								},
								{
									Name:     "locations",
									Strategy: "prepared-image",
									Args: map[string]string{
										"image": "y",
									},
								}},
						},
						Status: v1alpha1.SessionStatus{
							Refs: []*v1alpha1.RefStatus{
								{
									Ref: v1alpha1.Ref{
										Name:     "details",
										Strategy: "prepared-image",
										Args: map[string]string{
											"image": "x",
										}},
									Resources: []*v1alpha1.RefResource{{Kind: &kind, Name: &name, Action: &action}},
								},
								{
									Ref: v1alpha1.Ref{
										Name:     "ratings",
										Strategy: "telepresence",
									},
									Resources: []*v1alpha1.RefResource{{Kind: &kind, Name: &name, Action: &action}},
								},
								{
									Ref: v1alpha1.Ref{
										Name:     "locations",
										Strategy: "prepared-image",
										Args: map[string]string{
											"image": "x",
										}},
									Resources: []*v1alpha1.RefResource{{Kind: &kind, Name: &name, Action: &action}},
									Targets: []*v1alpha1.LabeledRefResource{
										{
											RefResource: v1alpha1.RefResource{
												Kind:   &kind,
												Name:   &locatedTargetName,
												Action: &locatedAction,
											},
										},
									},
								},
							},
						},
					},
				}
			})
			It("call revert when a status.ref.strategy differ from spec.ref.strategy", func() {
				locator.Action = foundTestLocator
				revertor.Action = removeResourceStatus("test", "details")
				res, err := controller.Reconcile(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				Expect(locator.WasCalled).To(BeTrue())
				Expect(mutator.WasCalled).To(BeTrue())
				Expect(revertor.WasCalled).To(BeTrue())

				modified := GetSession("test", "test-session")
				Expect(modified.Status).ToNot(BeNil())
				Expect(modified.Status.Refs).To(HaveLen(3))
			})
			Context("ensure updated spec.ref is reflected in status.ref", func() {
				It("when the strategy differ", func() {
					locator.Action = foundTestLocator
					revertor.Action = removeResourceStatus("test", "details")
					controller.Reconcile(req)

					modified := GetStatusRef("details", GetSession("test", "test-session"))
					Expect(modified).ToNot(BeNil())
					Expect(modified.Strategy).To(Equal("telepresence"))
					Expect(modified.Args).To(BeNil())
					Expect(modified.Resources).To(HaveLen(0))
				})
				It("when the args differ", func() {
					locator.Action = foundTestLocator
					revertor.Action = removeResourceStatus("test", "details")
					controller.Reconcile(req)

					modified := GetStatusRef("ratings", GetSession("test", "test-session"))
					Expect(modified).ToNot(BeNil())
					Expect(modified.Strategy).To(Equal("prepared-image"))
					Expect(modified.Args).To(Equal(map[string]string{"image": "x"}))
					Expect(modified.Resources).To(HaveLen(0))
				})
				It("when the same args differ", func() {
					locator.Action = foundTestLocator
					revertor.Action = removeResourceStatus("test", "details")
					controller.Reconcile(req)

					modified := GetStatusRef("locations", GetSession("test", "test-session"))
					Expect(modified).ToNot(BeNil())
					Expect(modified.Strategy).To(Equal("prepared-image"))
					Expect(modified.Args).To(Equal(map[string]string{"image": "y"}))
					Expect(modified.Resources).To(HaveLen(0))
				})
			})
			Context("ensure targets are reflected on update", func() {
				It("when any change happen", func() {
					locator.Action = foundTestLocatorTarget("test-a", "test-b")
					revertor.Action = noOp
					controller.Reconcile(req)

					modified := GetStatusRef("locations", GetSession("test", "test-session"))
					Expect(modified).ToNot(BeNil())
					Expect(modified.Targets).To(HaveLen(2))
				})
			})
		})
	})
	Context("session deletion", func() {
		BeforeEach(func() {
			objects = []runtime.Object{
				&v1alpha1.Session{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "test-session",
						Namespace:         "test",
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
						Finalizers:        []string{session.Finalizer},
					},
					Spec: v1alpha1.SessionSpec{
						Refs: []v1alpha1.Ref{{Name: "details"}},
					},
					Status: v1alpha1.SessionStatus{
						Refs: []*v1alpha1.RefStatus{
							{
								Ref:       v1alpha1.Ref{Name: "details"},
								Resources: []*v1alpha1.RefResource{{Kind: &kind, Name: &name, Action: &action}},
							},
						},
					},
				},
			}
		})
		It("revertors call when session removed", func() {
			locator.Action = foundTestLocator
			revertor.Action = removeResourceStatus("test", "details")
			res, err := controller.Reconcile(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Requeue).To(BeFalse())

			Expect(locator.WasCalled).To(BeFalse())
			Expect(mutator.WasCalled).To(BeFalse())
			Expect(revertor.WasCalled).To(BeTrue())
		})
		It("status removed when session removed", func() {
			locator.Action = foundTestLocator
			revertor.Action = removeResourceStatus("test", "details")
			res, err := controller.Reconcile(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Requeue).To(BeFalse())

			modified := GetSession("test", "test-session")
			Expect(modified.Status).ToNot(BeNil())
			Expect(modified.Status.Refs).To(HaveLen(0))
		})
		It("finalizer removed", func() {
			res, err := controller.Reconcile(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Requeue).To(BeFalse())

			modified := GetSession("test", "test-session")
			Expect(modified.ObjectMeta.Finalizers).To(HaveLen(0))
		})
	})
})

// notFound Action for Locator tracker.
func notFoundTestLocator(ctx model.SessionContext, ref *model.Ref) bool {
	return false
}

// found Action for Locator tracker.
func foundTestLocator(ctx model.SessionContext, ref *model.Ref) bool {
	return true
}

// found Action for Locator tracker.
func foundTestLocatorTarget(names ...string) func(ctx model.SessionContext, ref *model.Ref) bool {
	return func(ctx model.SessionContext, ref *model.Ref) bool {
		for _, name := range names {
			ref.AddTargetResource(model.NewLocatedResource("test", name, map[string]string{}))
		}
		return true
	}
}

// noOp Action for Mutator/Revertor trackers.
func noOp(ctx model.SessionContext, ref *model.Ref) error {
	return nil
}

type trackedLocator struct {
	WasCalled bool
	Action    model.Locator
}

func (t *trackedLocator) Do(ctx model.SessionContext, ref *model.Ref) bool {
	t.WasCalled = true
	return t.Action(ctx, ref)
}

// addResource Action for mutator tracker.
func addResourceStatus(status model.ResourceStatus) func(ctx model.SessionContext, ref *model.Ref) error {
	return func(ctx model.SessionContext, ref *model.Ref) error {
		ref.AddResourceStatus(status)
		return nil
	}
}

type trackedMutator struct {
	WasCalled bool
	Action    model.Mutator
}

func (t *trackedMutator) Do(ctx model.SessionContext, ref *model.Ref) error {
	t.WasCalled = true
	return t.Action(ctx, ref)
}

// removeResource Action for revertor tracker.
func removeResourceStatus(kind, name string) func(ctx model.SessionContext, ref *model.Ref) error { //nolint:unparam //reason kind is always receiving 'test' so far
	return func(ctx model.SessionContext, ref *model.Ref) error {
		ref.RemoveResourceStatus(model.ResourceStatus{Kind: kind, Name: name})
		return nil
	}
}

type trackedRevertor struct {
	WasCalled bool
	Action    model.Revertor
}

func (t *trackedRevertor) Do(ctx model.SessionContext, ref *model.Ref) error {
	t.WasCalled = true
	return t.Action(ctx, ref)
}
