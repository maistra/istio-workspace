package k8s_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/maistra/istio-workspace/pkg/k8s"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model/new"
	"github.com/maistra/istio-workspace/pkg/openshift"
)

var _ = Describe("Operations for k8s Service kind", func() {

	var objects []runtime.Object
	var ctx new.SessionContext

	CreateTestRef := func() new.Ref {
		return new.Ref{
			KindName:  new.RefKindName{Name: "test-ref"},
			Namespace: "test",
			Strategy:  "telepresence",
			Args:      map[string]string{"version": "0.103"},
		}
	}
	CreateTestLocatorStore := func(kind string, labels map[string]string) new.LocatorStore {
		l := new.LocatorStore{}
		l.Report(new.LocatorStatus{Kind: kind, Name: "test-ref", Labels: labels, Action: new.ActionLocated})

		return l
	}
	JustBeforeEach(func() {
		schema := runtime.NewScheme()
		err := corev1.AddToScheme(schema)
		Expect(err).ToNot(HaveOccurred())
		ctx = new.SessionContext{
			Context:   context.Background(),
			Name:      "test",
			Namespace: "test",
			Log:       log.CreateOperatorAwareLogger("test").WithValues("type", "k8s-service"),
			Client:    fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(objects...).Build(),
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
			ref := CreateTestRef()
			store := CreateTestLocatorStore(k8s.DeploymentKind, map[string]string{"app": "not found"})
			k8s.ServiceLocator(ctx, ref, store.Store, store.Report)
			Expect(store.Store(k8s.ServiceKind)).To(HaveLen(0))
		})

		It("should report true on found", func() {
			ref := CreateTestRef()
			store := CreateTestLocatorStore(k8s.DeploymentKind, map[string]string{"app": "x"})
			k8s.ServiceLocator(ctx, ref, store.Store, store.Report)
			Expect(store.Store(k8s.ServiceKind)).To(HaveLen(1))
		})

		It("should find services for Deployment", func() {
			ref := CreateTestRef()
			store := CreateTestLocatorStore(k8s.DeploymentKind, map[string]string{"app": "x"})
			k8s.ServiceLocator(ctx, ref, store.Store, store.Report)
			services := store.Store(k8s.ServiceKind)
			Expect(len(services)).To(Equal(1))

			Expect(services[0].Name).To(Equal("test-1"))
		})
		It("should find services for DeploymentConfig", func() {
			ref := CreateTestRef()
			store := CreateTestLocatorStore(openshift.DeploymentConfigKind, map[string]string{"app": "x"})
			k8s.ServiceLocator(ctx, ref, store.Store, store.Report)
			services := store.Store(k8s.ServiceKind)
			Expect(len(services)).To(Equal(1))

			Expect(services[0].Name).To(Equal("test-1"))
		})
		It("should return service hostname", func() {
			ref := CreateTestRef()
			store := CreateTestLocatorStore(k8s.DeploymentKind, map[string]string{"app": "x"})
			k8s.ServiceLocator(ctx, ref, store.Store, store.Report)
			hosts := new.GetTargetHostNames(store.Store)
			Expect(len(hosts)).To(Equal(1))

			Expect(hosts[0].Name).To(Equal("test-1"))
		})
		It("should add all matching services", func() {
			ref := CreateTestRef()
			store := CreateTestLocatorStore(k8s.DeploymentKind, map[string]string{"app": "z"})
			k8s.ServiceLocator(ctx, ref, store.Store, store.Report)
			services := store.Store(k8s.ServiceKind)
			Expect(len(services)).To(Equal(2))

			getNames := func(list []new.LocatorStatus) []string {
				var names []string
				for _, l := range list {
					names = append(names, l.Name)
				}

				return names
			}
			Expect(getNames(services)).To(ConsistOf("test-2", "test-3"))
		})

	})
})
