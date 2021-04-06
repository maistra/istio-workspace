package session_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	istiov1alpha1 "github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	testclient "github.com/maistra/istio-workspace/pkg/client/clientset/versioned/fake"
	"github.com/maistra/istio-workspace/pkg/internal/session"
)

var _ = Describe("Session Client operations", func() {

	sampleSession := &istiov1alpha1.Session{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "maistra.io/v1alpha1",
			Kind:       "Session",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "sample-session",
		},
		Spec: istiov1alpha1.SessionSpec{
			Refs: []istiov1alpha1.Ref{
				{Name: "test-details"},
			},
		},
	}

	Context("Session creation", func() {

		fakeClient := testclient.NewSimpleClientset()
		client, _ := session.NewClient(fakeClient, "test-namespace")

		It("should get created session by its name", func() {
			creationErr := client.Create(sampleSession)
			Expect(creationErr).ToNot(HaveOccurred())

			foundSession, getErr := client.Get("sample-session")
			Expect(getErr).ToNot(HaveOccurred())

			Expect(foundSession.Name).To(Equal(sampleSession.Name))
		})

	})

	Context("Session deletion", func() {

		fakeClient := testclient.NewSimpleClientset()
		client, _ := session.NewClient(fakeClient, "test-namespace")

		BeforeEach(func() {
			creationErr := client.Create(sampleSession)
			Expect(creationErr).ToNot(HaveOccurred())
		})

		It("should delete session by its name", func() {
			deleteErr := client.Delete(sampleSession)
			Expect(deleteErr).ToNot(HaveOccurred())

			_, getErr := client.Get("sample-session")
			Expect(getErr).To(HaveOccurred())

			var statusError *k8serrors.StatusError
			Expect(errors.As(getErr, &statusError)).To(BeTrue())

			Expect(statusError.Status().Code).To(Equal(int32(404)))
		})

	})
})
