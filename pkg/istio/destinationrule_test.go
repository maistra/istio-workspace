package istio_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	istionetworkv1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/istio"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/test/testclient"
)

var _ = Describe("Operations for istio DestinationRule kind", func() {

	GetName := func(s *istionetworkv1alpha3.Subset) string { return s.Name }

	var (
		objects []runtime.Object
		c       client.Client
		ctx     model.SessionContext
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
		ctx = model.SessionContext{
			Name:      "test",
			Namespace: "test",
			Client:    c,
			Log:       log.CreateOperatorAwareLogger("destinationrule"),
		}
	})

	Context("mutators", func() {

		Context("existing rule", func() {

			var (
				ref *model.Ref
			)

			BeforeEach(func() {
				ref = &model.Ref{
					KindName: model.ParseRefKindName("customer-v1"),
					Targets: []model.LocatedResourceStatus{
						model.NewLocatedResource("Deployment", "customer-v1", map[string]string{"version": "v1"}),
						model.NewLocatedResource("Service", "customer-mutate", nil),
					},
				}
			})

			It("should add reference", func() {
				err := istio.DestinationRuleMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				dr := get.DestinationRules("test", testclient.HasRefPredicate)
				Expect(dr.Items).To(HaveLen(1))
			})

			It("should have one subset defined", func() {
				err := istio.DestinationRuleMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				dr := get.DestinationRules("test", testclient.HasRefPredicate)
				Expect(dr.Items[0].Spec.Subsets).To(HaveLen(1))
			})

			It("should have one subset defined with name", func() {
				err := istio.DestinationRuleMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				dr := get.DestinationRules("test", testclient.HasRefPredicate)
				Expect(dr.Items[0].Spec.Subsets).To(ContainElement(WithTransform(GetName, Equal(ref.GetNewVersion(ctx.Name)))))
			})

			It("should not create new destination rules for subsequents mutations", func() {
				err := istio.DestinationRuleMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				// apply twice
				err = istio.DestinationRuleMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				dr := get.DestinationRules("test", testclient.HasRefPredicate)
				Expect(dr.Items).To(HaveLen(1))
				Expect(dr.Items[0].Spec.Subsets).To(HaveLen(1))
				Expect(dr.Items[0].Spec.Subsets).To(ContainElement(WithTransform(GetName, Equal(ref.GetNewVersion(ctx.Name)))))
			})

			It("should keep traficpolicy from target", func() {
				err := istio.DestinationRuleMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				dr := get.DestinationRules("test", testclient.HasRefPredicate)
				Expect(dr.Items[0].Spec.Subsets[0].TrafficPolicy).ToNot(BeNil())
				Expect(dr.Items[0].Spec.Subsets[0].TrafficPolicy.ConnectionPool.Http.MaxRetries).To(Equal(int32(100)))
			})
		})

		Context("missing rule", func() {

			It("should fail when no rules found", func() {
				ref := &model.Ref{
					KindName: model.ParseRefKindName("customer-v5"),
					Targets: []model.LocatedResourceStatus{
						model.NewLocatedResource("Deployment", "customer-v5", map[string]string{"version": "v5"}),
						model.NewLocatedResource("Service", "customer-missing", nil),
					},
				}
				err := istio.DestinationRuleMutator(ctx, ref)
				Expect(err).To(HaveOccurred())
				fmt.Println(err)
				Expect(err.Error()).To(ContainSubstring("failed finding subset with given host and version"))
			})
		})
	})

	Context("revertors", func() {

		var (
			ref *model.Ref
		)

		BeforeEach(func() {
			ref = &model.Ref{
				KindName: model.ParseRefKindName("customer-v1"),
				Targets: []model.LocatedResourceStatus{
					model.NewLocatedResource("Deployment", "customer-v1", map[string]string{"version": "v1"}),
					model.NewLocatedResource("Service", "customer-other", nil),
				},
			}
		})

		Context("existing rule", func() {

			It("should remove reference", func() {
				err := istio.DestinationRuleMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				err = istio.DestinationRuleRevertor(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				dr := get.DestinationRules("test", testclient.HasRefPredicate)
				Expect(dr.Items).To(BeEmpty())
			})

			It("should not fail on subsequent remove of reference", func() {
				err := istio.DestinationRuleMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				err = istio.DestinationRuleRevertor(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				err = istio.DestinationRuleRevertor(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				dr := get.DestinationRules("test", testclient.HasRefPredicate)
				Expect(dr.Items).To(BeEmpty())
			})

		})
	})
})
