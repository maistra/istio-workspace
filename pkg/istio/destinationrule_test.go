package istio_test

import (
	"fmt"

	"github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/istio"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/test/testclient"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	istionetworkv1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Operations for istio DestinationRule kind", func() {

	GetName := func(s *istionetworkv1alpha3.Subset) string { return s.Name }

	var (
		objects []runtime.Object
		c       client.Client
		ctx     model.SessionContext
		get     *testclient.Getters
	)

	const namespace = "test"

	BeforeEach(func() {
		objects = []runtime.Object{
			&istionetwork.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "customer-mutate",
					Namespace: namespace,
				},
				Spec: istionetworkv1alpha3.DestinationRule{
					Host: "customer-mutate",
					Subsets: []*istionetworkv1alpha3.Subset{
						{
							Name: "v1",
							Labels: map[string]string{
								"version": "v1",
							},
							TrafficPolicy: &istionetworkv1alpha3.TrafficPolicy{
								ConnectionPool: &istionetworkv1alpha3.ConnectionPoolSettings{
									Http: &istionetworkv1alpha3.ConnectionPoolSettings_HTTPSettings{
										MaxRetries: 100,
									},
								},
							},
						},
					},
					TrafficPolicy: &istionetworkv1alpha3.TrafficPolicy{
						ConnectionPool: &istionetworkv1alpha3.ConnectionPoolSettings{
							Http: &istionetworkv1alpha3.ConnectionPoolSettings_HTTPSettings{
								MaxRetries: 10,
							},
						},
					},
				},
			},
			&istionetwork.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "customer-other",
					Namespace: namespace,
				},
				Spec: istionetworkv1alpha3.DestinationRule{
					Host: "customer-other",
					Subsets: []*istionetworkv1alpha3.Subset{
						{
							Name: "v1",
							Labels: map[string]string{
								"version": "v1",
							},
						},
					},
				},
			},
		}
	})

	JustBeforeEach(func() {
		schema, _ := v1alpha1.SchemeBuilder.Register(
			&istionetwork.DestinationRule{},
			&istionetwork.DestinationRuleList{}).Build()

		c = fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(objects...).Build()
		get = testclient.New(c)
		ctx = model.SessionContext{
			Name:      namespace,
			Namespace: namespace,
			Client:    c,
			Log:       log.CreateOperatorAwareLogger("destinationrule"),
		}
	})

	Context("locators", func() {

		var (
			ref          model.Ref
			locators     model.LocatorStore
			modificators model.ModificatorStore
		)

		BeforeEach(func() {
			ref = model.Ref{
				KindName:  model.ParseRefKindName("customer-v1"),
				Namespace: namespace,
			}
			locators = createLocatorStore()

			modificators = model.ModificatorStore{}
		})

		It("should trigger create action when reference is created", func() {
			// when
			err := istio.DestinationRuleLocator(ctx, ref, locators.Store, locators.Report)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(locators.Store(istio.DestinationRuleKind)).To(HaveLen(1))
			dr := locators.Store(istio.DestinationRuleKind)[0]
			Expect(dr.Action).To(Equal(model.ActionCreate))
			Expect(dr.Name).To(Equal("customer-other"))
		})

		It("should trigger delete and create action when reference is updated", func() {
			// given
			err := istio.DestinationRuleLocator(ctx, ref, locators.Store, locators.Report)
			Expect(err).ToNot(HaveOccurred())
			istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
			Expect(modificators.Stored).To(HaveLen(1))

			// when
			ref.Strategy = "prepared-image"
			newLocatorStore := createLocatorStore()
			err = istio.DestinationRuleLocator(ctx, ref, newLocatorStore.Store, newLocatorStore.Report)
			Expect(err).ToNot(HaveOccurred())

			// then
			actions := newLocatorStore.Store(istio.DestinationRuleKind)
			Expect(actions).To(HaveLen(2))
			Expect(actions[0].Action).To(Equal(model.ActionDelete))
			Expect(actions[0].Name).To(Equal("dr-customer-v1-customer-other-test"))
			Expect(actions[1].Action).To(Equal(model.ActionCreate))
			Expect(actions[1].Name).To(Equal("customer-other"))
		})

		It("should trigger revert action when reference is removed", func() {
			// given
			err := istio.DestinationRuleLocator(ctx, ref, locators.Store, locators.Report)
			Expect(err).ToNot(HaveOccurred())
			istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
			Expect(modificators.Stored).To(HaveLen(1))

			// when
			ref.Remove = true
			newLocatorStore := createLocatorStore()
			err = istio.DestinationRuleLocator(ctx, ref, newLocatorStore.Store, newLocatorStore.Report)
			Expect(err).ToNot(HaveOccurred())

			// then
			actions := newLocatorStore.Store(istio.DestinationRuleKind)
			Expect(actions).To(HaveLen(1))
			Expect(actions[0].Action).To(Equal(model.ActionDelete))
			Expect(actions[0].Name).To(Equal("dr-customer-v1-customer-other-test"))
		})

	})

	Context("modificators", func() {

		Context("existing rule", func() {

			var (
				ref          model.Ref
				locators     model.LocatorStore
				modificators model.ModificatorStore
			)

			BeforeEach(func() {
				ref = model.Ref{
					KindName:  model.ParseRefKindName("customer-v1"),
					Namespace: namespace,
				}
				locators = createLocatorStore()
				locators.Report(model.LocatorStatus{Resource: model.Resource{Kind: "DestinationRule", Namespace: namespace, Name: "customer-mutate"}, Action: model.ActionCreate})
				modificators = model.ModificatorStore{}
			})

			It("should add reference", func() {
				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				dr := get.DestinationRules(namespace, testclient.HasRefPredicate)
				Expect(dr.Items).To(HaveLen(1))
			})

			It("should have one subset defined", func() {
				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				dr := get.DestinationRules(namespace, testclient.HasRefPredicate)
				Expect(dr.Items[0].Spec.Subsets).To(HaveLen(1))
			})

			It("should have one subset defined with name", func() {
				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				dr := get.DestinationRules(namespace, testclient.HasRefPredicate)
				Expect(dr.Items[0].Spec.Subsets).To(ContainElement(WithTransform(GetName, Equal(model.GetCreatedVersion(locators.Store, ctx.Name)))))
			})

			It("should not create new destination rules for subsequents mutations", func() {
				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				// apply twice
				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(2))
				fmt.Println(modificators.Stored)

				dr := get.DestinationRules(namespace, testclient.HasRefPredicate)
				Expect(dr.Items).To(HaveLen(1))
				Expect(dr.Items[0].Spec.Subsets).To(HaveLen(1))
				Expect(dr.Items[0].Spec.Subsets).To(ContainElement(WithTransform(GetName, Equal(model.GetCreatedVersion(locators.Store, ctx.Name)))))
			})

			It("should keep traffic policy from target", func() {
				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				dr := get.DestinationRules(namespace, testclient.HasRefPredicate)
				Expect(dr.Items[0].Spec.Subsets[0].TrafficPolicy).ToNot(BeNil())
				Expect(dr.Items[0].Spec.Subsets[0].TrafficPolicy.ConnectionPool.Http.MaxRetries).To(Equal(int32(100)))
			})
		})

		Context("missing rule", func() {

			// https://github.com/maistra/istio-workspace/issues/856
			XIt("should fail when no rules found", func() {
				ref := model.Ref{
					KindName:  model.ParseRefKindName("customer-v5"),
					Namespace: namespace,
				}
				locators := model.LocatorStore{}
				locators.Report(model.LocatorStatus{Resource: model.Resource{Kind: "Deployment", Namespace: namespace, Name: "customer-v3"}, Labels: map[string]string{"version": "v5"}})
				locators.Report(model.LocatorStatus{Resource: model.Resource{Kind: "Service", Namespace: namespace, Name: "customer-missing"}})
				modificators := model.ModificatorStore{}

				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())
				Expect(modificators.Stored[0].Error).To(ContainSubstring("failed finding subset with given host and version"))
			})
		})
	})

	Context("revertors", func() {

		var (
			ref          model.Ref
			locators     model.LocatorStore
			modificators model.ModificatorStore
		)

		BeforeEach(func() {
			ref = model.Ref{
				KindName:  model.ParseRefKindName("customer-v1"),
				Namespace: namespace,
			}
			locators = model.LocatorStore{}
			locators.Report(model.LocatorStatus{Resource: model.Resource{Kind: "Deployment", Namespace: namespace, Name: "customer-v1"}, Labels: map[string]string{"version": "v1"}})
			locators.Report(model.LocatorStatus{Resource: model.Resource{Kind: "Service", Namespace: namespace, Name: "customer-other"}})
			locators.Report(model.LocatorStatus{Resource: model.Resource{Kind: "DestinationRule", Namespace: namespace, Name: "customer-other"}, Action: model.ActionDelete})
			modificators = model.ModificatorStore{}

		})

		Context("existing rule", func() {

			It("should remove reference", func() {
				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(2))

				dr := get.DestinationRules(namespace, testclient.HasRefPredicate)
				Expect(dr.Items).To(BeEmpty())
			})

			It("should not fail on subsequent remove of reference", func() {
				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).ToNot(HaveLen(1))

				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).ToNot(HaveLen(1))

				dr := get.DestinationRules(namespace, testclient.HasRefPredicate)
				Expect(dr.Items).To(BeEmpty())
			})

		})
	})
})

func createLocatorStore() model.LocatorStore {
	locators := model.LocatorStore{}
	locators.Report(model.LocatorStatus{Resource: model.Resource{Kind: "Deployment", Namespace: "test", Name: "customer-v1"}, Labels: map[string]string{"version": "v1"}})
	locators.Report(model.LocatorStatus{Resource: model.Resource{Kind: "Service", Namespace: "test", Name: "customer-other"}})

	return locators
}
