package session_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/controllers/session"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/test/testclient"
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
		get        *testclient.Getters
	)
	JustBeforeEach(func() {
		locator = &trackedLocator{Action: notFoundTestLocator}
		mutator = &trackedMutator{Action: noOpModifier}
		manipulators := session.Manipulators{
			Locators: []model.Locator{locator.Do},
			Handlers: []model.ModificatorRegistrar{
				func() (client.Object, model.Modificator) {
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
		get = testclient.New(c)
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

				modified := get.Session("test", "test-session")
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
				mutator.Action = reportSuccess()

				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				modified := get.Session("test", "test-session")
				Expect(modified.Status).ToNot(BeNil())
				Expect(modified.Status.Conditions).To(HaveLen(1))
				Expect(modified.Status.Conditions[0].Source.Name).To(Equal("test"))
			})
			It("should update status with the corresponding route", func() {
				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				modified := get.Session("test", "test-session")
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
								{Source: v1alpha1.Source{
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
				mutator.Action = reportSuccess()

				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				Expect(locator.WasCalled).To(BeTrue())
				Expect(mutator.WasCalled).To(BeTrue())
				Expect(mutator.Refs[0].Remove).To(BeFalse())
			})
			It("should update existing status when new mutation occurs", func() {
				locator.Action = foundTestLocatorTarget("details2")
				mutator.Action = reportSuccess()

				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				modified := get.Session("test", "test-session")
				Expect(modified.Status).ToNot(BeNil())
				Expect(modified.Status.Conditions).To(HaveLen(2))

				getNames := func(list []*v1alpha1.Condition) []string {
					var names []string
					for _, l := range list {
						names = append(names, l.Source.Name)
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
								{Source: v1alpha1.Source{
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
				mutator.Action = reportSuccess()
				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				Expect(locator.WasCalled).To(BeTrue())
				Expect(mutator.WasCalled).To(BeTrue())
				Expect(mutator.Refs[0].Remove).To(BeTrue())
			})

			It("should remove status when ref is removed", func() {
				locator.Action = foundTestLocator
				mutator.Action = reportSuccess()
				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				modified := get.Session("test", "test-session")
				Expect(modified.Status).ToNot(BeNil())
				Expect(modified.Status.Conditions).To(HaveLen(0))
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
									Source: v1alpha1.Source{
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
				mutator.Action = reportSuccess()
				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				Expect(locator.WasCalled).To(BeTrue())
				Expect(mutator.WasCalled).To(BeTrue())
				Expect(mutator.Refs[0].Remove).To(BeTrue())
			})

			It("should remove finalizer", func() {
				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				// Then - object should have been finalized/deleted if no finalizers present
				_, err = get.SessionWithError("test", "test-session")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

// notFound Action for Locator tracker.
func notFoundTestLocator(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.LocatorStatusReporter) error {
	return nil
}

// found Action for Locator tracker.
func foundTestLocator(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.LocatorStatusReporter) error {
	report(model.LocatorStatus{Resource: model.Resource{Kind: "X", Name: "test"}, Action: model.ActionCreate})

	return nil
}

// found Action for Locator tracker.
func foundTestLocatorTarget(names ...string) func(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.LocatorStatusReporter) error {
	return func(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.LocatorStatusReporter) error {
		for _, name := range names {
			report(model.LocatorStatus{Resource: model.Resource{Kind: "X", Name: name}, Action: model.ActionCreate})
		}

		return nil
	}
}

// noOpModifier Action for Mutator/Revertor trackers.
func noOpModifier(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.ModificatorStatusReporter) {
}

type trackedLocator struct {
	WasCalled bool
	Action    model.Locator
}

func (t *trackedLocator) Do(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.LocatorStatusReporter) error {
	t.WasCalled = true

	return t.Action(ctx, ref, store, report)
}

// reportSuccess Action for mutator tracker.
func reportSuccess() func(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.ModificatorStatusReporter) {
	return func(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.ModificatorStatusReporter) {
		for _, l := range store() {
			report(model.ModificatorStatus{LocatorStatus: l, Success: true, Error: nil})
		}
	}
}

type trackedMutator struct {
	WasCalled bool
	Action    model.Modificator
	Refs      []model.Ref
}

func (t *trackedMutator) Do(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.ModificatorStatusReporter) {
	t.WasCalled = true
	t.Refs = append(t.Refs, ref)

	t.Action(ctx, ref, store, report)
}
