package k8s_test

import (
	"context"

	"github.com/maistra/istio-workspace/pkg/k8s"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/openshift"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Operations for k8s Service kind", func() {

	var objects []runtime.Object
	var ctx model.SessionContext

	CreateTestRef := func(kind string, lables map[string]string) model.Ref {
		return model.Ref{
			Name:      "test-ref",
			Namespace: "test",
			Strategy:  "telepresence",
			Targets:   []model.LocatedResourceStatus{model.NewLocatedResource(kind, "test-ref", lables)},
			Args:      map[string]string{"version": "0.103"},
		}
	}
	
	JustBeforeEach(func() {
		schema := runtime.NewScheme()
		err := corev1.AddToScheme(schema)
		Expect(err).ToNot(HaveOccurred())
		ctx = model.SessionContext{
			Context:   context.Background(),
			Name:      "test",
			Namespace: "test",
			Log:       log.CreateOperatorAwareLogger("test").WithValues("type", "k8s-service"),
			Client:    fake.NewFakeClientWithScheme(schema, objects...),
		}
	})

	Context("locators", func() {
		BeforeEach(func() {
			objects = []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-1",
						Namespace: "test",
					},
					Spec: corev1.ServiceSpec{
						Selector: map[string]string{
							"app": "x",
						},
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-2",
						Namespace: "test",
					},
					Spec: corev1.ServiceSpec{
						Selector: map[string]string{
							"app": "z",
						},
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-3",
						Namespace: "test",
					},
					Spec: corev1.ServiceSpec{
						Selector: map[string]string{
							"app": "z",
						},
					},
				},
			}
		})

		It("should report false on not found", func() {
			ref := CreateTestRef(k8s.DeploymentKind, map[string]string{"app": "not found"})
			locatorErr := k8s.ServiceLocator(ctx, &ref)
			Expect(locatorErr).To(BeFalse())
		})

		It("should report true on found", func() {
			ref := CreateTestRef(k8s.DeploymentKind, map[string]string{"app": "x"})
			locatorErr := k8s.ServiceLocator(ctx, &ref)
			Expect(locatorErr).To(BeTrue())
		})

		It("should find services for Deployment", func() {
			ref := CreateTestRef(k8s.DeploymentKind, map[string]string{"app": "x"})
			k8s.ServiceLocator(ctx, &ref)
			services := ref.GetTargetsByKind(k8s.ServiceKind)
			Expect(len(services)).To(Equal(1))

			Expect(services[0].Name).To(Equal("test-1"))
		})
		It("should find services for DeploymentConfig", func() {
			ref := CreateTestRef(openshift.DeploymentConfigKind, map[string]string{"app": "x"})
			k8s.ServiceLocator(ctx, &ref)
			services := ref.GetTargetsByKind(k8s.ServiceKind)
			Expect(len(services)).To(Equal(1))

			Expect(services[0].Name).To(Equal("test-1"))
		})
		It("should return service hostname", func() {
			ref := CreateTestRef(k8s.DeploymentKind, map[string]string{"app": "x"})
			k8s.ServiceLocator(ctx, &ref)
			hosts := ref.GetTargetHostNames()
			Expect(len(hosts)).To(Equal(1))

			Expect(hosts[0].Name).To(Equal("test-1"))
		})

		It("should add all matching services", func() {
			ref := CreateTestRef(k8s.DeploymentKind, map[string]string{"app": "z"})
			k8s.ServiceLocator(ctx, &ref)
			services := ref.GetTargetsByKind(k8s.ServiceKind)
			Expect(len(services)).To(Equal(2))

			Expect(services[0].Name).To(Equal("test-2"))
			Expect(services[1].Name).To(Equal("test-3"))
		})

	})
})
