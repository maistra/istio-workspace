package k8s_test

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/maistra/istio-workspace/pkg/k8s"
	"github.com/maistra/istio-workspace/pkg/model"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Operations for k8s Deployment kind", func() {

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
			locatorErr := k8s.DeploymentLocator(ctx, &ref)
			Expect(locatorErr).To(BeFalse())
		})

		It("should report true on found", func() {
			ref := model.Ref{Name: "test-ref"}
			locatorErr := k8s.DeploymentLocator(ctx, &ref)
			Expect(locatorErr).To(BeTrue())
		})

	})

	Context("mutators", func() {

		BeforeEach(func() {
			objects = []runtime.Object{
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-ref",
						Namespace: "test",
						Labels: map[string]string{
							"version": "0.0.1",
						},
					},

					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{},
						},
						Template: v1.PodTemplateSpec{
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

		It("should add suffix to the cloned deployment", func() {
			ref := model.Ref{Name: "test-ref"}
			mutatorErr := k8s.DeploymentMutator(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			deployment := appsv1.Deployment{}
			err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: ctx.Namespace, Name: ref.Name + "-" + ctx.Name}, &deployment)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("telepresence mutations", func() {

			It("should change container to telepresence", func() {
				ref := model.Ref{Name: "test-ref"}
				mutatorErr := k8s.DeploymentMutator(ctx, &ref)
				Expect(mutatorErr).ToNot(HaveOccurred())

				deployment := appsv1.Deployment{}
				err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: ctx.Namespace, Name: ref.Name + "-" + ctx.Name}, &deployment)
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(ContainSubstring("datawire/telepresence-k8s:"))
			})

			It("should change add required env variables", func() {
				ref := model.Ref{Name: "test-ref"}
				mutatorErr := k8s.DeploymentMutator(ctx, &ref)
				Expect(mutatorErr).ToNot(HaveOccurred())

				deployment := appsv1.Deployment{}
				err := ctx.Client.Get(ctx, types.NamespacedName{Namespace: ctx.Namespace, Name: ref.Name + "-" + ctx.Name}, &deployment)
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.Spec.Template.Spec.Containers[0].Env[0].Name).To(Equal("TELEPRESENCE_CONTAINER_NAMESPACE"))
				Expect(deployment.Spec.Template.Spec.Containers[0].Env[0].ValueFrom).ToNot(BeNil())
			})

		})

	})

	Context("revertors", func() {

		BeforeEach(func() {
			objects = []runtime.Object{
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-ref",
						Namespace: "test",
						Labels: map[string]string{
							"version": "0.0.1",
						},
					},

					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{},
						},
						Template: v1.PodTemplateSpec{
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

		It("should revert to original deployment", func() {
			ref := model.Ref{Name: "test-ref"}
			mutatorErr := k8s.DeploymentMutator(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			deployment := appsv1.Deployment{}

			mutatedFetchErr := ctx.Client.Get(ctx, types.NamespacedName{Namespace: ctx.Namespace, Name: ref.Name + "-" + ctx.Name}, &deployment)
			Expect(mutatedFetchErr).ToNot(HaveOccurred())

			revertorErr := k8s.DeploymentRevertor(ctx, &ref)
			Expect(revertorErr).ToNot(HaveOccurred())

			revertedFetchErr := ctx.Client.Get(ctx, types.NamespacedName{Namespace: ctx.Namespace, Name: ref.Name + "-" + ctx.Name}, &deployment)
			Expect(revertedFetchErr).To(HaveOccurred())
		})

	})
})
