package istio

import (
	"github.com/maistra/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/test/operator"

	. "github.com/onsi/ginkgo"
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
		get     *operator.Helpers
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
						&istionetworkv1alpha3.Subset{
							Name: "v1",
							Labels: map[string]string{
								"version": "v1",
							},
						},
					},
				},
			},
			&istionetwork.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "customer-revert",
					Namespace: "test",
				},
				Spec: istionetworkv1alpha3.DestinationRule{
					Host: "customer-revert",
					Subsets: []*istionetworkv1alpha3.Subset{
						&istionetworkv1alpha3.Subset{
							Name: "v1",
							Labels: map[string]string{
								"version": "v1",
							},
						},
						&istionetworkv1alpha3.Subset{
							Name: "dr-test",
							Labels: map[string]string{
								"version": "dr-test",
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

		c = fake.NewFakeClientWithScheme(schema, objects...)
		get = operator.New(c)
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
					Name: "customer-v1",
					Targets: []model.LocatedResourceStatus{
						model.NewLocatedResource("Deployment", "customer-v1", map[string]string{"version": "v1"}),
						model.NewLocatedResource("Service", "customer-mutate", nil),
					},
				}
			})

			It("new subset added", func() {
				err := DestinationRuleMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				dr := get.DestinationRule("test", "customer-mutate")
				Expect(dr.Spec.Subsets).To(HaveLen(2))
			})

			It("new subset added with name", func() {
				err := DestinationRuleMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				dr := get.DestinationRule("test", "customer-mutate")
				Expect(dr.Spec.Subsets).To(ContainElement(WithTransform(GetName, Equal("v1-test"))))
			})

			It("new subset only added once", func() {
				err := DestinationRuleMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				// apply twice
				err = DestinationRuleMutator(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				dr := get.DestinationRule("test", "customer-mutate")
				Expect(dr.Spec.Subsets).To(HaveLen(2))
				Expect(dr.Spec.Subsets).To(ContainElement(WithTransform(GetName, Equal("v1-test"))))
			})
		})
	})

	Context("revertors", func() {

		var (
			ref *model.Ref
		)

		BeforeEach(func() {
			ref = &model.Ref{
				Name: "customer-v1",
				Targets: []model.LocatedResourceStatus{
					model.NewLocatedResource("Deployment", "customer-v1", map[string]string{"version": "v1"}),
					model.NewLocatedResource("Service", "customer-revert", nil),
				},
			}
		})

		Context("existing rule", func() {

			It("new subset removed", func() {
				err := DestinationRuleRevertor(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				dr := get.DestinationRule("test", "customer-revert")
				Expect(dr.Spec.Subsets).To(HaveLen(2))
			})

			It("correct subset removed", func() {
				err := DestinationRuleRevertor(ctx, ref)
				Expect(err).ToNot(HaveOccurred())

				dr := get.DestinationRule("test", "customer-revert")
				Expect(dr.Spec.Subsets).ToNot(ContainElement(WithTransform(GetName, Equal("v1-test"))))
			})
		})
	})
})
