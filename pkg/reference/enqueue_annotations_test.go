package reference_test

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/maistra/istio-workspace/pkg/reference"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

	It("should fail to add when missing namespace", func() {
		// when
		err := reference.Add(types.NamespacedName{Name: "session1"}, deployment)
		Expect(err).To((HaveOccurred()))
	})

	It("should fail to add when missing name", func() {
		// when
		err := reference.Add(types.NamespacedName{Namespace: "test"}, deployment)
		Expect(err).To((HaveOccurred()))
	})

	It("should add single reference when single session exist", func() {
		// when
		reference.Add(types.NamespacedName{Namespace: "test", Name: "session1"}, deployment)

		// then
		Expect(deployment.Annotations[reference.NamespacedNameAnnotation]).To(Equal("test/session1"))

	})

	It("should add multiple reference when multiple session exist", func() {
		// when
		reference.Add(types.NamespacedName{Namespace: "test", Name: "session1"}, deployment)
		reference.Add(types.NamespacedName{Namespace: "test", Name: "session2"}, deployment)

		// then
		Expect(deployment.Annotations[reference.NamespacedNameAnnotation]).To(Equal("test/session1,test/session2"))
	})

	It("should prevent from having duplicate annotations", func() {
		// when
		reference.Add(types.NamespacedName{Namespace: "test", Name: "session1"}, deployment)
		reference.Add(types.NamespacedName{Namespace: "test", Name: "session2"}, deployment)
		reference.Add(types.NamespacedName{Namespace: "test", Name: "session1"}, deployment)

		// then
		Expect(deployment.Annotations[reference.NamespacedNameAnnotation]).To(Equal("test/session1,test/session2"))
	})

	It("should fail to remove when missing namespace", func() {
		// when
		err := reference.Remove(types.NamespacedName{Name: "session1"}, deployment)
		Expect(err).To(HaveOccurred())
	})
	It("should fail to remove when missing name", func() {
		// when
		err := reference.Remove(types.NamespacedName{Namespace: "test"}, deployment)
		Expect(err).To(HaveOccurred())
	})

	It("should remove reference when no reference exist", func() {
		// given
		reference.Add(types.NamespacedName{Namespace: "test", Name: "session1"}, deployment)

		// when
		reference.Remove(types.NamespacedName{Namespace: "test", Name: "session1"}, deployment)

		// then
		Expect(deployment.Annotations[reference.NamespacedNameAnnotation]).To(BeEmpty())
	})

	It("should remove single reference when multiple reference exist", func() {
		// given
		reference.Add(types.NamespacedName{Namespace: "test", Name: "session1"}, deployment)
		reference.Add(types.NamespacedName{Namespace: "test", Name: "session2"}, deployment)

		// when
		reference.Remove(types.NamespacedName{Namespace: "test", Name: "session1"}, deployment)

		// then
		Expect(deployment.Annotations[reference.NamespacedNameAnnotation]).To(Equal("test/session2"))
	})

	It("should get references when one reference exist", func() {
		// given
		Expect(reference.Get(deployment)).To(HaveLen(0))
		reference.Add(types.NamespacedName{Namespace: "test", Name: "session1"}, deployment)

		// then
		typeNames := reference.Get(deployment)
		Expect(typeNames).To(HaveLen(1))
		Expect(typeNames[0].String()).To(Equal("test/session1"))
	})

	It("should get references when multiple reference exist", func() {
		// given
		Expect(reference.Get(deployment)).To(HaveLen(0))
		reference.Add(types.NamespacedName{Namespace: "test", Name: "session1"}, deployment)
		reference.Add(types.NamespacedName{Namespace: "test", Name: "session2"}, deployment)

		// then
		typeNames := reference.Get(deployment)
		Expect(typeNames).To(HaveLen(2))
		Expect(typeNames[0].String()).To(Equal("test/session1"))
		Expect(typeNames[1].String()).To(Equal("test/session2"))
	})
})
