package k8s_test

import (
	"context"

	"github.com/maistra/istio-workspace/pkg/k8s"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/test/operator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Operations for k8s Deployment kind", func() {

	var (
		objects []runtime.Object
		c       client.Client
		ctx     model.SessionContext
		get     *operator.Helpers
	)

	CreateTestRef := func() model.Ref {
		return model.Ref{
			Name:     "test-ref",
			Strategy: "telepresence",
			Targets:  []model.LocatedResourceStatus{model.NewLocatedResource(k8s.DeploymentKind, "test-ref", map[string]string{"version": "v1"})},
			Args:     map[string]string{"version": "0.103"},
		}
	}
	JustBeforeEach(func() {
		schema := runtime.NewScheme()
		err := appsv1.AddToScheme(schema)
		Expect(err).ToNot(HaveOccurred())
		c = fake.NewFakeClientWithScheme(schema, objects...)
		get = operator.New(c)
		ctx = model.SessionContext{
			Context:   context.Background(),
			Name:      "test",
			Namespace: "test",
			Log:       log.CreateOperatorAwareLogger("test").WithValues("type", "k8s-deployment"),
			Client:    c,
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
					TypeMeta: metav1.TypeMeta{
						Kind: "Deployment",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-ref",
						Namespace: "test",
						Labels: map[string]string{
							"version": "0.0.1",
						},
						CreationTimestamp: metav1.Now(),
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
										LivenessProbe: &v1.Probe{
											Handler: v1.Handler{
												HTTPGet: &v1.HTTPGetAction{
													Path: "/healthz",
													Port: intstr.FromInt(9080),
												},
											},
										},
										ReadinessProbe: &v1.Probe{
											Handler: v1.Handler{
												HTTPGet: &v1.HTTPGetAction{
													Path: "/healthz",
													Port: intstr.FromInt(9080),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
		})

		It("should add suffix to the cloned deployment", func() {
			ref := CreateTestRef()
			mutatorErr := k8s.DeploymentMutator(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			_ = get.Deployment(ctx.Namespace, ref.Name+"-v1-"+ctx.Name)
		})

		It("should remove liveness probe from cloned deployment", func() {
			ref := CreateTestRef()
			mutatorErr := k8s.DeploymentMutator(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			deployment := get.Deployment(ctx.Namespace, ref.Name+"-v1-"+ctx.Name)
			Expect(deployment.Spec.Template.Spec.Containers[0].LivenessProbe).To(BeNil())
		})

		It("should remove readiness probe from cloned deployment", func() {
			ref := CreateTestRef()
			mutatorErr := k8s.DeploymentMutator(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			deployment := get.Deployment(ctx.Namespace, ref.Name+"-v1-"+ctx.Name)
			Expect(deployment.Spec.Template.Spec.Containers[0].ReadinessProbe).To(BeNil())
		})

		It("should update selector", func() {
			ref := CreateTestRef()
			mutatorErr := k8s.DeploymentMutator(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			deployment := get.Deployment(ctx.Namespace, ref.Name+"-v1-"+ctx.Name)
			Expect(deployment.Spec.Template.Spec.Containers[0].ReadinessProbe).To(BeNil())
			Expect(deployment.Spec.Selector.MatchLabels["version"]).To(BeEquivalentTo("v1-test"))
		})

		It("should only mutate if Target is of kind Deployment", func() {
			notMatchingRef := model.Ref{Name: "test-ref", Targets: []model.LocatedResourceStatus{model.NewLocatedResource("Service", "test-ref", nil)}}
			mutatorErr := k8s.DeploymentMutator(ctx, &notMatchingRef)
			Expect(mutatorErr).ToNot(HaveOccurred())

			_, err := get.DeploymentWithError(ctx.Namespace, notMatchingRef.Name+"-v1-"+ctx.Name)
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		Context("telepresence mutations", func() {

			It("should change container to telepresence", func() {
				ref := CreateTestRef()
				mutatorErr := k8s.DeploymentMutator(ctx, &ref)
				Expect(mutatorErr).ToNot(HaveOccurred())

				deployment := get.Deployment(ctx.Namespace, ref.Name+"-v1-"+ctx.Name)
				Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(ContainSubstring("datawire/telepresence-k8s:"))
			})

			It("should change add required env variables", func() {
				ref := CreateTestRef()
				mutatorErr := k8s.DeploymentMutator(ctx, &ref)
				Expect(mutatorErr).ToNot(HaveOccurred())

				deployment := get.Deployment(ctx.Namespace, ref.Name+"-v1-"+ctx.Name)
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
						CreationTimestamp: metav1.Now(),
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
			ref := CreateTestRef()
			mutatorErr := k8s.DeploymentMutator(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			_, mutatedFetchErr := get.DeploymentWithError(ctx.Namespace, ref.Name+"-v1-"+ctx.Name)
			Expect(mutatedFetchErr).ToNot(HaveOccurred())

			revertorErr := k8s.DeploymentRevertor(ctx, &ref)
			Expect(revertorErr).ToNot(HaveOccurred())

			_, revertedFetchErr := get.DeploymentWithError(ctx.Namespace, ref.Name+"-v1-"+ctx.Name)
			Expect(revertedFetchErr).To(HaveOccurred())
			Expect(errors.IsNotFound(revertedFetchErr)).To(BeTrue())
		})

	})
})
