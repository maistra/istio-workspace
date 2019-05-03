package openshift_test

import (
	"context"

	"github.com/aslakknutsen/istio-workspace/pkg/model"
	"github.com/aslakknutsen/istio-workspace/pkg/openshift"

	appsv1 "github.com/openshift/api/apps/v1"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Operations for openshift DeploymentConfig kind", func() {

	var objects []runtime.Object
	var ctx model.SessionContext
	JustBeforeEach(func() {
		schema := runtime.NewScheme()
		err := appsv1.AddToScheme(schema)
		Expect(err).ToNot(HaveOccurred())
		ctx = model.SessionContext{
			Context:   context.TODO(),
			Name:      "test",
			Namespace: "test",
			Log:       logf.Log.WithName("test"),
			Client:    fake.NewFakeClientWithScheme(schema, objects...),
		}
	})

	Context("locators", func() {
		BeforeEach(func() {
			objects = []runtime.Object{
				&appsv1.DeploymentConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-ref",
						Namespace: "test",
					},
				},
			}
		})

		It("should report false on not found", func() {
			ref := model.Ref{Name: "test-ref-other"}
			locatorErr := openshift.DeploymentConfigLocator(ctx, &ref)
			Expect(locatorErr).To(BeFalse())
		})

		It("should report true on found", func() {
			ref := model.Ref{Name: "test-ref"}
			locatorErr := openshift.DeploymentConfigLocator(ctx, &ref)
			Expect(locatorErr).To(BeTrue())
		})

	})

	Context("mutators", func() {

		BeforeEach(func() {
			objects = []runtime.Object{
				&appsv1.DeploymentConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-ref",
						Namespace: "test",
						Labels: map[string]string{
							"version": "0.0.1",
						},
					},
					Spec: appsv1.DeploymentConfigSpec{
						Selector: map[string]string{"A": "A"},
						Template: &v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "datawire/hello-world:latest",
										Env:   []v1.EnvVar{},
									},
								},
							},
						},
					},
				},
			}
		})

		It("should add suffix to the cloned deploymentconfig", func() {
			ref := model.Ref{Name: "test-ref"}
			mutatorErr := openshift.DeploymentConfigMutator(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			deployment := appsv1.DeploymentConfig{}
			err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: ctx.Namespace, Name: ref.Name + "-test"}, &deployment)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("telepresence mutations", func() {

			It("should change container to telepresence", func() {
				ref := model.Ref{Name: "test-ref"}
				mutatorErr := openshift.DeploymentConfigMutator(ctx, &ref)
				Expect(mutatorErr).ToNot(HaveOccurred())

				deployment := appsv1.DeploymentConfig{}
				err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: ctx.Namespace, Name: ref.Name + "-test"}, &deployment)
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(ContainSubstring("datawire/telepresence-k8s:0.98"))
			})

			It("should change add required env variables", func() {
				ref := model.Ref{Name: "test-ref"}
				mutatorErr := openshift.DeploymentConfigMutator(ctx, &ref)
				Expect(mutatorErr).ToNot(HaveOccurred())

				deployment := appsv1.DeploymentConfig{}
				err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: ctx.Namespace, Name: ref.Name + "-test"}, &deployment)
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.Spec.Template.Spec.Containers[0].Env[0].Name).To(Equal("TELEPRESENCE_CONTAINER_NAMESPACE"))
				Expect(deployment.Spec.Template.Spec.Containers[0].Env[0].ValueFrom).ToNot(BeNil())
			})

		})

	})

	Context("revertors", func() {

		BeforeEach(func() {
			objects = []runtime.Object{
				&appsv1.DeploymentConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-ref",
						Namespace: "test",
						Labels: map[string]string{
							"version": "0.0.1",
						},
					},
					Spec: appsv1.DeploymentConfigSpec{
						Selector: map[string]string{"A": "A"},
						Template: &v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "datawire/hello-world:latest",
										Env:   []v1.EnvVar{},
									},
								},
							},
						},
					},
				},
			}
		})

		It("should revert to original deploymentconfig", func() {
			ref := model.Ref{Name: "test-ref"}
			mutatorErr := openshift.DeploymentConfigMutator(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			deployment := appsv1.DeploymentConfig{}

			mutatedFetchErr := ctx.Client.Get(ctx, types.NamespacedName{Namespace: ctx.Namespace, Name: ref.Name + "-test"}, &deployment)
			Expect(mutatedFetchErr).ToNot(HaveOccurred())

			revertorErr := openshift.DeploymentConfigRevertor(ctx, &ref)
			Expect(revertorErr).ToNot(HaveOccurred())

			revertedFetchErr := ctx.Client.Get(ctx, types.NamespacedName{Namespace: ctx.Namespace, Name: ref.Name + "-test"}, &deployment)
			Expect(revertedFetchErr).To(HaveOccurred())
		})

	})
})
