package k8s_test

import (
	"context"

	"github.com/aslakknutsen/istio-workspace/pkg/k8s"
	"github.com/aslakknutsen/istio-workspace/pkg/model"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Operations for k8 Deployment kind", func() {

	var objects []runtime.Object
	var ctx model.SessionContext
	JustBeforeEach(func() {
		ctx = model.SessionContext{
			Context:   context.TODO(),
			Name:      "test",
			Namespace: "test",
			Log:       logf.Log.WithName("test"),
			Client:    fake.NewFakeClient(objects...),
		}
	})

	Context("locators", func() {
		BeforeEach(func() {
			objects = []runtime.Object{
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-ref",
						Namespace: "test",
					},
				},
			}
		})

		It("should report false on not found", func() {
			ref := model.Ref{Name: "test-ref-other"}

			Expect(k8s.DeploymentLocator(ctx, &ref)).To(Equal(false))
		})
		XIt("should report false on other found", func() {
			ref := model.Ref{Name: "test-ref"}

			Expect(k8s.DeploymentLocator(ctx, &ref)).To(Equal(false))
		})
		It("should report true on found", func() {
			ref := model.Ref{Name: "test-ref"}

			Expect(k8s.DeploymentLocator(ctx, &ref)).To(Equal(true))
		})

	})

	Context("mutators", func() {
		It("should fail invoking develop cmd when telepresence binary is not on $PATH", func() {
		})

	})

	Context("revertors", func() {

		It("should fail invoking develop cmd when telepresence binary is not on $PATH", func() {
		})

	})
})
