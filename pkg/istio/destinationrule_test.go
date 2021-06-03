package istio_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	istionetworkv1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/istio"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model/new"
	"github.com/maistra/istio-workspace/pkg/reference"
	"github.com/maistra/istio-workspace/test/testclient"
)

var _ = Describe("Operations for istio DestinationRule kind", func() {

	GetName := func(s *istionetworkv1alpha3.Subset) string { return s.Name }

	var (
		objects []runtime.Object
		c       client.Client
		ctx     new.SessionContext
		get     *testclient.Getters
	)

	BeforeEach(func() {
		objects = []runtime.Object{
			&istionetwork.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "customer-mutate",
					Namespace: "test",
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
					Namespace: "test",
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
		ctx = new.SessionContext{
			Name:      "test",
			Namespace: "test",
			Client:    c,
			Log:       log.CreateOperatorAwareLogger("destinationrule"),
		}
	})

	//TODO: Missing Locators tests, Modificator tests need DestinationRule locatorstatus

	Context("mutators", func() {

		Context("existing rule", func() {

			var (
				ref          new.Ref
				locators     new.LocatorStore
				modificators new.ModificatorStore
			)

			BeforeEach(func() {
				ref = new.Ref{
					KindName: new.ParseRefKindName("customer-v1"),
				}
				locators = new.LocatorStore{}
				locators.Report(new.LocatorStatus{Kind: "Deployment", Namespace: "test", Name: "customer-v1", Labels: map[string]string{"version": "v1"}})
				locators.Report(new.LocatorStatus{Kind: "Service", Namespace: "test", Name: "customer-mutate"})
				locators.Report(new.LocatorStatus{Kind: "DestinationRule", Namespace: "test", Name: "customer-mutate", Action: new.ActionCreate})
				modificators = new.ModificatorStore{}
			})

			It("should add reference", func() {
				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				dr := get.DestinationRules("test", testclient.HasRefPredicate)
				Expect(dr.Items).To(HaveLen(1))
			})

			It("should have one subset defined", func() {
				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				dr := get.DestinationRules("test", testclient.HasRefPredicate)
				Expect(dr.Items[0].Spec.Subsets).To(HaveLen(1))
			})

			It("should have one subset defined with name", func() {
				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				dr := get.DestinationRules("test", testclient.HasRefPredicate)
				Expect(dr.Items[0].Spec.Subsets).To(ContainElement(WithTransform(GetName, Equal(new.GetNewVersion(locators.Store, ctx.Name)))))
			})

			It("should not create new destination rules for subsequents mutations", func() {
				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				// apply twice
				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(2))
				fmt.Println(modificators.Stored)

				dr := get.DestinationRules("test", testclient.HasRefPredicate)
				Expect(dr.Items).To(HaveLen(1))
				Expect(dr.Items[0].Spec.Subsets).To(HaveLen(1))
				Expect(dr.Items[0].Spec.Subsets).To(ContainElement(WithTransform(GetName, Equal(new.GetNewVersion(locators.Store, ctx.Name)))))
			})

			It("should keep traficpolicy from target", func() {
				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				dr := get.DestinationRules("test", testclient.HasRefPredicate)
				Expect(dr.Items[0].Spec.Subsets[0].TrafficPolicy).ToNot(BeNil())
				Expect(dr.Items[0].Spec.Subsets[0].TrafficPolicy.ConnectionPool.Http.MaxRetries).To(Equal(int32(100)))
			})
		})

		Context("missing rule", func() {

			// TODO: This should be moved to overall Validation of Locators result, possible simulate Found but Deleted before attempt to mutate?
			XIt("should fail when no rules found", func() {
				ref := new.Ref{
					KindName: new.ParseRefKindName("customer-v5"),
				}
				locators := new.LocatorStore{}
				locators.Report(new.LocatorStatus{Kind: "Deployment", Namespace: "test", Name: "customer-v3", Labels: map[string]string{"version": "v5"}})
				locators.Report(new.LocatorStatus{Kind: "Service", Namespace: "test", Name: "customer-missing"})
				modificators := new.ModificatorStore{}

				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())
				Expect(modificators.Stored[0].Error).To(ContainSubstring("failed finding subset with given host and version"))
			})
		})
	})

	Context("revertors", func() {

		var (
			ref          new.Ref
			locators     new.LocatorStore
			modificators new.ModificatorStore
		)

		BeforeEach(func() {
			ref = new.Ref{
				KindName: new.ParseRefKindName("customer-v1"),
			}
			locators = new.LocatorStore{}
			locators.Report(new.LocatorStatus{Kind: "Deployment", Namespace: "test", Name: "customer-v1", Labels: map[string]string{"version": "v1"}})
			locators.Report(new.LocatorStatus{Kind: "Service", Namespace: "test", Name: "customer-other"})
			locators.Report(new.LocatorStatus{Kind: "DestinationRule", Namespace: "test", Name: "customer-other", Action: new.ActionDelete})
			modificators = new.ModificatorStore{}

		})

		Context("existing rule", func() {

			It("should remove reference", func() {
				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(1))
				Expect(modificators.Stored[0].Error).ToNot(HaveOccurred())

				istio.DestinationRuleModificator(ctx, ref, locators.Store, modificators.Report)
				Expect(modificators.Stored).To(HaveLen(2))

				dr := get.DestinationRules("test", testclient.HasRefPredicate)
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

				dr := get.DestinationRules("test", testclient.HasRefPredicate)
				Expect(dr.Items).To(BeEmpty())
			})

		})
	})
})

func TestX(t *testing.T) {
	kubeCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	restCfg, err := kubeCfg.ClientConfig()
	if err != nil {
		fmt.Println(err)
	}

	schema := runtime.NewScheme()
	istionetwork.AddToScheme(schema)

	c, err := client.New(restCfg, client.Options{Scheme: schema})
	if err != nil {
		fmt.Println(err)
	}

	ctx := new.SessionContext{Client: c, Context: context.Background()}

	resources, err := istio.GetDestinationRules(ctx, "aslak-test", reference.Match("test"))
	if err != nil {

		fmt.Println(err)
	}

	fmt.Println(resources)
}
