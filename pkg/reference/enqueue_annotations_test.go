package reference_test

import (
	"github.com/maistra/istio-workspace/pkg/reference"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Enqueue Annotations", func() {

	var deployment *appsv1.Deployment

	BeforeEach(func() {
		deployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-ref",
				Namespace: "test",
			},
		}
	})

	It("should add single reference when single session exist", func() {

		reference.SetOwnerAnnotations(types.NamespacedName{Namespace: "test", Name: "session1"}, deployment)

		Expect(deployment.Annotations[reference.NamespacedNameAnnotation]).To((Equal("test/session1")))

	})
	PIt("should add multiple reference when multiple session exist", func() {})
	PIt("should remove single reference when multiple session exist", func() {})
	PIt("should remove reference when no session exist", func() {})

})
