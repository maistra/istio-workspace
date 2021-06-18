package session_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/controllers/session"
	"github.com/maistra/istio-workspace/pkg/model/new"
)

var kind, name = "X", "details"

var _ = Describe("Basic session manipulation", func() {
	var (
		objects    []runtime.Object
		controller reconcile.Reconciler
		req        reconcile.Request
		schema     *runtime.Scheme
		c          client.Client
		locator    *trackedLocator
		mutator    *trackedMutator
	)
	GetSession := func(c *client.Client) func(namespace, name string) v1alpha1.Session {
		return func(namespace, name string) v1alpha1.Session {
			s := v1alpha1.Session{}
			err := (*c).Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &s)
			Expect(err).ToNot(HaveOccurred())

			return s
		}
	}(&c)
	/*
		GetStatusRef := func(name string, session v1alpha1.Session) []v1alpha1.Condition {
			var conditions []v1alpha1.Condition

			for _, condition := range session.Status.Conditions {
				if condition.Target.Name == name {
					conditions = append(conditions, *condition)
				}
			}

			return conditions
		}
	*/
	JustBeforeEach(func() {
		locator = &trackedLocator{Action: notFoundTestLocator}
		mutator = &trackedMutator{Action: noOpModifier}
		manipulators := session.Manipulators{
			Locators: []new.Locator{locator.Do},
			Handlers: []new.ModificatorRegistrar{
				func() (client.Object, new.Modificator) {
					return nil, mutator.Do
				},
			},
		}

		schema, _ = v1alpha1.SchemeBuilder.Build()
		req = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      "test-session",
				Namespace: "test",
			},
		}
		c = fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(objects...).Build()
		controller = session.NewStandaloneReconciler(c, manipulators)
	})

	Context("session creation", func() {

		Context("session not found", func() {
			BeforeEach(func() {
				objects = []runtime.Object{}
			})

			It("should not retry", func() {
				res, err := controller.Reconcile(context.Background(), req)
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

			It("should add finalizer", func() {
				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				modified := GetSession("test", "test-session")
				Expect(modified.ObjectMeta.Finalizers).To(HaveLen(1))
			})

			It("should mutate when target is located", func() {
				locator.Action = foundTestLocator

				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				Expect(locator.WasCalled).To(BeTrue())
				Expect(mutator.WasCalled).To(BeTrue())
			})
			It("should update the status when mutation occurs", func() {
				locator.Action = foundTestLocator
				mutator.Action = addResourceStatus(true, nil)

				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				modified := GetSession("test", "test-session")
				Expect(modified.Status).ToNot(BeNil())
				Expect(modified.Status.Conditions).To(HaveLen(1))
				Expect(modified.Status.Conditions[0].Target.Name).To(Equal("test"))
			})
			It("should update status with the corresponding route", func() {
				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				modified := GetSession("test", "test-session")
				fmt.Println(modified.Status.Route)
				Expect(modified.Status.Route).ToNot(BeNil())
				Expect(modified.Status.Route.Type).To(Equal(session.RouteStrategyHeader))
				Expect(modified.Status.Route.Name).To(Equal(session.DefaultRouteHeaderName))
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
							Conditions: []*v1alpha1.Condition{
								{Target: v1alpha1.Target{
									Name: name,
									Kind: kind,
									Ref:  "details",
								}},
							},
						},
					},
				}
			})
			It("should not call revertors when mutation occurs", func() {
				locator.Action = foundTestLocator
				mutator.Action = addResourceStatus(true, nil)

				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				Expect(locator.WasCalled).To(BeTrue())
				Expect(mutator.WasCalled).To(BeTrue())
				Expect(mutator.Refs[0].Deleted).To(BeFalse())
			})
			It("should update existing status when new mutation occurs", func() {
				locator.Action = foundTestLocatorTarget("details2")
				mutator.Action = addResourceStatus(true, nil)

				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				modified := GetSession("test", "test-session")
				Expect(modified.Status).ToNot(BeNil())
				Expect(modified.Status.Conditions).To(HaveLen(2))

				getNames := func(list []*v1alpha1.Condition) []string {
					var names []string
					for _, l := range list {
						names = append(names, l.Target.Name)
					}

					return names
				}
				Expect(getNames(modified.Status.Conditions)).To(ConsistOf("details2", "details2"))
				Expect(*modified.Status.State).To(Equal(v1alpha1.StateSuccess))
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
							Conditions: []*v1alpha1.Condition{
								{Target: v1alpha1.Target{
									Name: name,
									Kind: kind,
									Ref:  "details",
								}},
							},
						},
					},
				}
			})
			It("should call revertors when ref is removed", func() {
				locator.Action = foundTestLocator
				mutator.Action = addResourceStatus(true, nil)
				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				Expect(locator.WasCalled).To(BeTrue())
				Expect(mutator.WasCalled).To(BeTrue())
				Expect(mutator.Refs[0].Deleted).To(BeTrue())
			})

			It("should remove status when ref is removed", func() {
				locator.Action = foundTestLocator
				mutator.Action = addResourceStatus(true, nil)
				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				modified := GetSession("test", "test-session")
				Expect(modified.Status).ToNot(BeNil())
				Expect(modified.Status.Conditions).To(HaveLen(0))
			})
			Context("updated reference", func() {
				BeforeEach(func() {
					// TODO: missing a way to detect that a ref has changed, e.g. new strategy...
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
								Conditions: []*v1alpha1.Condition{
									{
										Target: v1alpha1.Target{
											Name: "test",
											Kind: "X",
											Ref:  "details",
										},
									},
									{
										Target: v1alpha1.Target{
											Name: "test",
											Kind: "X",
											Ref:  "ratings",
										},
									},
									{
										Target: v1alpha1.Target{
											Name: "locations",
											Kind: "X",
											Ref:  "details",
										},
									},
								},
							},
						},
					}
				})
				It("should call revert when a status.ref.strategy differs from spec.ref.strategy", func() {
					locator.Action = foundTestLocator
					mutator.Action = addResourceStatus(true, nil)
					res, err := controller.Reconcile(context.Background(), req)
					Expect(err).ToNot(HaveOccurred())
					Expect(res.Requeue).To(BeFalse())

					Expect(locator.WasCalled).To(BeTrue())
					Expect(mutator.WasCalled).To(BeTrue())

					modified := GetSession("test", "test-session")
					Expect(modified.Status).ToNot(BeNil())
					Expect(modified.Status.Conditions).To(HaveLen(3))
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
							Refs: []v1alpha1.Ref{},
						},
						Status: v1alpha1.SessionStatus{
							Conditions: []*v1alpha1.Condition{
								{
									Target: v1alpha1.Target{
										Name: "test",
										Kind: "X",
										Ref:  "details",
									},
								},
							},
						},
					},
				}
			})

			It("should call revertors when session removed", func() {
				locator.Action = foundTestLocator
				mutator.Action = addResourceStatus(true, nil)
				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				Expect(locator.WasCalled).To(BeTrue())
				Expect(mutator.WasCalled).To(BeTrue())
				Expect(mutator.Refs[0].Deleted).To(BeTrue())
			})

			It("should remove status when session removed", func() {
				locator.Action = foundTestLocator
				mutator.Action = addResourceStatus(true, nil)
				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				modified := GetSession("test", "test-session")
				Expect(modified.Status).ToNot(BeNil())
				Expect(modified.Status.Conditions).To(HaveLen(0))
			})

			It("should remove finalizer", func() {
				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				modified := GetSession("test", "test-session")
				Expect(modified.ObjectMeta.Finalizers).To(HaveLen(0))
			})
		})
	})
})

// notFound Action for Locator tracker.
func notFoundTestLocator(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.LocatorStatusReporter) error {
	return nil
}

// found Action for Locator tracker.
func foundTestLocator(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.LocatorStatusReporter) error {
	report(new.LocatorStatus{Kind: "X", Name: "test", Action: new.ActionCreate})

	return nil
}

// found Action for Locator tracker.
func foundTestLocatorTarget(names ...string) func(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.LocatorStatusReporter) error {
	return func(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.LocatorStatusReporter) error {
		for _, name := range names {
			report(new.LocatorStatus{Kind: "X", Name: name, Action: new.ActionCreate})
		}

		return nil
	}
}

// noOpModifier Action for Mutator/Revertor trackers.
func noOpModifier(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.ModificatorStatusReporter) {
}

type trackedLocator struct {
	WasCalled bool
	Action    new.Locator
}

func (t *trackedLocator) Do(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.LocatorStatusReporter) error {
	t.WasCalled = true

	return t.Action(ctx, ref, store, report)
}

// addResource Action for mutator tracker.
func addResourceStatus(success bool, err error) func(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.ModificatorStatusReporter) {
	return func(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.ModificatorStatusReporter) {
		for _, l := range store() {
			report(new.ModificatorStatus{LocatorStatus: l, Success: success, Error: err})
		}
	}
}

type trackedMutator struct {
	WasCalled bool
	Action    new.Modificator
	Refs      []new.Ref
}

func (t *trackedMutator) Do(ctx new.SessionContext, ref new.Ref, store new.LocatorStatusStore, report new.ModificatorStatusReporter) {
	t.WasCalled = true
	t.Refs = append(t.Refs, ref)

	t.Action(ctx, ref, store, report)
}
