package session

import (
	"context"
	"time"

	"github.com/aslakknutsen/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/aslakknutsen/istio-workspace/pkg/model"
	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Basic session manipulation", func() {
	var (
		objects                    []runtime.Object
		controller                 ReconcileSession
		req                        reconcile.Request
		schema                     *runtime.Scheme
		locator, mutator, revertor = &trackedLocator{Action: notFoundTestLocator}, &trackedMutator{Action: emptyTestMutator}, &trackedRevertor{Action: emptyTestRevertor}
	)
	GetClient := func(c *ReconcileSession) func() client.Client { return func() client.Client { return c.client } }(&controller)
	GetSession := func(c func() client.Client) func(namespace, name string) v1alpha1.Session {
		return func(namespace, name string) v1alpha1.Session {
			s := v1alpha1.Session{}
			err := c().Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, &s)
			Expect(err).ToNot(HaveOccurred())
			return s
		}
	}(GetClient)

	JustBeforeEach(func() {
		manipulators := Manipulators{
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
		controller = ReconcileSession{
			client:       fake.NewFakeClientWithScheme(schema, objects...),
			manipulators: manipulators,
		}
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
							Refs: []string{"details"},
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

			It("found locator", func() {
				locator.Action = foundTestLocator

				res, err := controller.Reconcile(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				Expect(locator.WasCalled).To(BeTrue())
				Expect(mutator.WasCalled).To(BeTrue())
			})

			It("status created", func() {
				locator.Action = foundTestLocator
				mutator.Action = basicTestMutator(model.ResourceStatus{Name: "details", Kind: "test", Action: model.ActionCreated})

				res, err := controller.Reconcile(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				Expect(locator.WasCalled).To(BeTrue())
				Expect(mutator.WasCalled).To(BeTrue())
				Expect(revertor.WasCalled).ToNot(BeTrue())

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
							Finalizers: []string{finalizer},
						},

						Spec: v1alpha1.SessionSpec{
							Refs: []string{"details", "details2"},
						},
						Status: v1alpha1.SessionStatus{
							Refs: []*v1alpha1.RefStatus{
								&v1alpha1.RefStatus{
									Name: "details",
								},
							},
						},
					},
				}
			})
			It("status updated", func() {
				locator.Action = foundTestLocator
				mutator.Action = basicTestMutator(model.ResourceStatus{Name: "details2", Kind: "test", Action: model.ActionCreated})

				res, err := controller.Reconcile(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				Expect(locator.WasCalled).To(BeTrue())
				Expect(mutator.WasCalled).To(BeTrue())
				Expect(revertor.WasCalled).ToNot(BeTrue())

				modified := GetSession("test", "test-session")
				Expect(modified.Status).ToNot(BeNil())
				Expect(modified.Status.Refs).To(HaveLen(2))
				Expect(modified.Status.Refs[0].Name).To(Equal("details"))
				Expect(modified.Status.Refs[1].Name).To(Equal("details2"))
			})
		})
		Context("removed reference", func() {
			BeforeEach(func() {
				kind, name, action := "test", "details", "created"
				objects = []runtime.Object{
					&v1alpha1.Session{
						ObjectMeta: metav1.ObjectMeta{
							Name:       "test-session",
							Namespace:  "test",
							Finalizers: []string{finalizer},
						},

						Spec: v1alpha1.SessionSpec{
							Refs: []string{},
						},
						Status: v1alpha1.SessionStatus{
							Refs: []*v1alpha1.RefStatus{
								&v1alpha1.RefStatus{
									Name:      "details",
									Resources: []*v1alpha1.RefResource{{Kind: &kind, Name: &name, Action: &action}},
								},
							},
						},
					},
				}
			})
			It("status removed", func() {
				locator.Action = foundTestLocator
				revertor.Action = basicTestRevertor("test", "details")
				res, err := controller.Reconcile(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				Expect(locator.WasCalled).To(BeTrue())
				Expect(mutator.WasCalled).To(BeTrue())
				Expect(revertor.WasCalled).To(BeTrue())

				modified := GetSession("test", "test-session")
				Expect(modified.Status).ToNot(BeNil())
				Expect(modified.Status.Refs).To(HaveLen(0))
			})
		})
	})
	Context("session deletion", func() {
		BeforeEach(func() {
			kind, name, action := "test", "details", "created"
			objects = []runtime.Object{
				&v1alpha1.Session{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "test-session",
						Namespace:         "test",
						DeletionTimestamp: &metav1.Time{Time: time.Now()},
						Finalizers:        []string{finalizer},
					},

					Spec: v1alpha1.SessionSpec{
						Refs: []string{"details"},
					},
					Status: v1alpha1.SessionStatus{
						Refs: []*v1alpha1.RefStatus{
							&v1alpha1.RefStatus{
								Name:      "details",
								Resources: []*v1alpha1.RefResource{{Kind: &kind, Name: &name, Action: &action}},
							},
						},
					},
				},
			}
		})
		It("references removed", func() {
			locator.Action = foundTestLocator
			revertor.Action = basicTestRevertor("test", "details")
			res, err := controller.Reconcile(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Requeue).To(BeFalse())

			Expect(locator.WasCalled).To(BeTrue())
			Expect(mutator.WasCalled).To(BeTrue())
			Expect(revertor.WasCalled).To(BeTrue())

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

func notFoundTestLocator(ctx model.SessionContext, ref *model.Ref) bool {
	return false
}

func foundTestLocator(ctx model.SessionContext, ref *model.Ref) bool {
	return true
}

type trackedLocator struct {
	WasCalled bool
	Action    model.Locator
}

func (t *trackedLocator) Do(ctx model.SessionContext, ref *model.Ref) bool {
	t.WasCalled = true
	return t.Action(ctx, ref)
}

func emptyTestMutator(ctx model.SessionContext, ref *model.Ref) error {
	return nil
}

func basicTestMutator(status model.ResourceStatus) func(ctx model.SessionContext, ref *model.Ref) error {
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

func emptyTestRevertor(ctx model.SessionContext, ref *model.Ref) error {
	return nil
}

func basicTestRevertor(kind, name string) func(ctx model.SessionContext, ref *model.Ref) error {
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
