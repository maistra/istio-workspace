package dynclient_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta/testrestmapper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/maistra/istio-workspace/pkg/k8s/dynclient"
)

const groupName = "apiextensions.k8s.io"

var _ = Describe("Testing dynamic client", func() {

	var client dynclient.Client

	gv := schema.GroupVersion{Group: groupName, Version: "v1"}

	BeforeEach(func() {
		sessionCRD := &unstructured.Unstructured{}
		sessionCRD.SetUnstructuredContent(map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", gv.Group, gv.Version),
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "sessions.workspace.maistra.io",
			},
			"spec": map[string]interface{}{
				"names": map[string]interface{}{
					"kind":     "Session",
					"listKind": "SessionList",
					"plural":   "sessions",
					"singular": "session",
				},
			},
		})
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(gv, &apiextensions.CustomResourceDefinition{})
		client = dynclient.NewClient(
			dynamicfake.NewSimpleDynamicClient(scheme, sessionCRD),
			fake.NewSimpleClientset(sessionCRD),
			testrestmapper.TestOnlyStaticRESTMapper(scheme),
		)
	})

	Context("using defined operations", func() {

		It("should be able to create a new resource", func() {
			// given
			sampleCRD := &unstructured.Unstructured{}
			sampleCRD.SetUnstructuredContent(map[string]interface{}{
				"apiVersion": "apiextensions.k8s.io/v1",
				"kind":       "CustomResourceDefinition",
				"metadata": map[string]interface{}{
					"name": "sample.ike.io",
				},
				"spec": map[string]interface{}{
					"names": map[string]interface{}{
						"kind":     "Sample",
						"listKind": "SampleList",
						"plural":   "samples",
						"singular": "sample",
					},
				},
			})

			// when
			err := client.Create(sampleCRD)

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should be able to delete existing resource", func() {
			// given
			existingCRD := &unstructured.Unstructured{}
			existingCRD.SetUnstructuredContent(map[string]interface{}{
				"apiVersion": "apiextensions.k8s.io/v1",
				"kind":       "CustomResourceDefinition",
				"metadata": map[string]interface{}{
					"name": "sessions.workspace.maistra.io",
				},
			})

			// when
			err := client.Delete(existingCRD)

			// then
			Expect(err).ToNot(HaveOccurred())
		})

	})

	Context("using underlying dynamic client", func() {

		It("should find defined CRD", func() {
			// when
			crd, err := client.Dynamic().Resource(schema.GroupVersionResource{
				Group:    "apiextensions.k8s.io",
				Version:  "v1",
				Resource: "customresourcedefinitions",
			}).Get(context.Background(), "sessions.workspace.maistra.io", metav1.GetOptions{})

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(crd.GetKind()).To(Equal("CustomResourceDefinition"))
			Expect(crd.GetName()).To(Equal("sessions.workspace.maistra.io"))
		})

		It("should fail looking up non-existing CRD", func() {
			// when
			_, err := client.Dynamic().Resource(schema.GroupVersionResource{
				Group:    "apiextensions.k8s.io",
				Version:  "v1",
				Resource: "customresourcedefinitions",
			}).Get(context.Background(), "sessions.ike.io", metav1.GetOptions{})

			// then
			Expect(k8sErrors.IsNotFound(err)).To(BeTrue())
		})

	})

})
