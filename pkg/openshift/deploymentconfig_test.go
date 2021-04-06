package openshift_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/openshift"
	"github.com/maistra/istio-workspace/pkg/reference"
	"github.com/maistra/istio-workspace/pkg/template"
	"github.com/maistra/istio-workspace/test/testclient"
)

var _ = Describe("Operations for openshift DeploymentConfig kind", func() {

	var (
		objects []runtime.Object
		c       client.Client
		ctx     model.SessionContext
		get     *testclient.Getters
	)

	CreateTestRef := func() model.Ref {
		return model.Ref{
			Name:     "test-ref",
			Strategy: "telepresence",
			Targets:  []model.LocatedResourceStatus{model.NewLocatedResource(openshift.DeploymentConfigKind, "test-ref", map[string]string{"version": "v1"})},
			Args:     map[string]string{"version": "0.103"},
		}
	}
	JustBeforeEach(func() {
		schema := runtime.NewScheme()
		err := appsv1.Install(schema)
		Expect(err).ToNot(HaveOccurred())
		c = fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(objects...).Build()
		get = testclient.New(c)
		ctx = model.SessionContext{
			Context:   context.Background(),
			Name:      "test",
			Namespace: "test",
			Log:       log.CreateOperatorAwareLogger("test").WithValues("type", "openshift-deploymentconfig"),
			Client:    c,
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
					Spec: appsv1.DeploymentConfigSpec{
						Template: &v1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{},
							},
						},
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
					TypeMeta: metav1.TypeMeta{
						Kind: "DeploymentConfig",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-ref",
						Namespace: "test",
						Labels: map[string]string{
							"version": "0.0.1",
						},
						CreationTimestamp: metav1.Now(),
					},
					Spec: appsv1.DeploymentConfigSpec{
						Selector: map[string]string{"version": "0.0.1"},
						Template: &v1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"version": "0.0.1",
								},
							},
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

		It("should add reference to cloned deployment", func() {
			ref := CreateTestRef()
			mutatorErr := openshift.DeploymentConfigMutator(template.NewDefaultEngine())(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			dc := get.DeploymentConfig(ctx.Namespace, ref.Name+"-"+ref.GetNewVersion(ctx.Name))
			Expect(reference.Get(&dc)).To(HaveLen(1))
		})

		It("should add suffix to the cloned deploymentconfig", func() {
			ref := CreateTestRef()
			mutatorErr := openshift.DeploymentConfigMutator(template.NewDefaultEngine())(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			_ = get.DeploymentConfig(ctx.Namespace, ref.Name+"-"+ref.GetNewVersion(ctx.Name))
		})

		It("should remove liveness probe from cloned deployment", func() {
			ref := CreateTestRef()
			mutatorErr := openshift.DeploymentConfigMutator(template.NewDefaultEngine())(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			deployment := get.DeploymentConfig(ctx.Namespace, ref.Name+"-"+ref.GetNewVersion(ctx.Name))
			Expect(deployment.Spec.Template.Spec.Containers[0].LivenessProbe).To(BeNil())
		})

		It("should remove readiness probe from cloned deployment", func() {
			ref := CreateTestRef()
			mutatorErr := openshift.DeploymentConfigMutator(template.NewDefaultEngine())(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			deployment := get.DeploymentConfig(ctx.Namespace, ref.Name+"-"+ref.GetNewVersion(ctx.Name))
			Expect(deployment.Spec.Template.Spec.Containers[0].ReadinessProbe).To(BeNil())
		})

		It("should update selector", func() {
			ref := CreateTestRef()
			mutatorErr := openshift.DeploymentConfigMutator(template.NewDefaultEngine())(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			deployment := get.DeploymentConfig(ctx.Namespace, ref.Name+"-"+ref.GetNewVersion(ctx.Name))
			Expect(deployment.Spec.Selector["version"]).To(BeEquivalentTo(model.GetSha("v1") + "-test"))
		})

		It("should only mutate if Target is of kind DeploymentConfig", func() {
			notMatchingRef := model.Ref{Name: "test-ref", Targets: []model.LocatedResourceStatus{model.NewLocatedResource("Service", "test-ref", nil)}}
			mutatorErr := openshift.DeploymentConfigMutator(template.NewDefaultEngine())(ctx, &notMatchingRef)
			Expect(mutatorErr).ToNot(HaveOccurred())

			_, err := get.DeploymentConfigWithError(ctx.Namespace, notMatchingRef.Name+"-"+notMatchingRef.GetNewVersion(ctx.Name))
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("should recreate cloned DeploymentConfig if deleted externally", func() {
			// given a normal setup
			ref := CreateTestRef()
			mutatorErr := openshift.DeploymentConfigMutator(template.NewDefaultEngine())(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			deployment := get.DeploymentConfig(ctx.Namespace, ref.Name+"-"+ref.GetNewVersion(ctx.Name))
			Expect(deployment.Spec.Selector["version"]).To(BeEquivalentTo(model.GetSha("v1") + "-test"))

			// when DeploymentConfig is deleted
			c.Delete(ctx, &deployment)

			_, err := get.DeploymentConfigWithError(ctx.Namespace, ref.Name+"-"+ref.GetNewVersion(ctx.Name))
			Expect(err).To(HaveOccurred())

			// then it should be recreated on next reconcile
			mutatorErr = openshift.DeploymentConfigMutator(template.NewDefaultEngine())(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			deployment = get.DeploymentConfig(ctx.Namespace, ref.Name+"-"+ref.GetNewVersion(ctx.Name))
			Expect(deployment.Spec.Selector["version"]).To(BeEquivalentTo(model.GetSha("v1") + "-test"))
		})

		Context("telepresence mutation strategy", func() {

			It("should change container to telepresence", func() {
				ref := CreateTestRef()
				mutatorErr := openshift.DeploymentConfigMutator(template.NewDefaultEngine())(ctx, &ref)
				Expect(mutatorErr).ToNot(HaveOccurred())

				deployment := get.DeploymentConfig(ctx.Namespace, ref.Name+"-"+ref.GetNewVersion(ctx.Name))
				Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(ContainSubstring("datawire/telepresence-k8s:"))
			})

			It("should change add required env variables", func() {
				ref := CreateTestRef()
				mutatorErr := openshift.DeploymentConfigMutator(template.NewDefaultEngine())(ctx, &ref)
				Expect(mutatorErr).ToNot(HaveOccurred())

				deployment := get.DeploymentConfig(ctx.Namespace, ref.Name+"-"+ref.GetNewVersion(ctx.Name))
				Expect(deployment.Spec.Template.Spec.Containers[0].Env[0].Name).To(Equal("TELEPRESENCE_CONTAINER_NAMESPACE"))
				Expect(deployment.Spec.Template.Spec.Containers[0].Env[0].ValueFrom).ToNot(BeNil())
			})

		})

		Context("existing mutation strategy", func() {

			It("should not create a clone", func() {
				ref := CreateTestRef()
				ref.Strategy = model.StrategyExisting

				mutatorErr := openshift.DeploymentConfigMutator(template.NewDefaultEngine())(ctx, &ref)
				Expect(mutatorErr).ToNot(HaveOccurred())

				_, err := get.DeploymentConfigWithError(ctx.Namespace, ref.Name+"-"+ref.GetNewVersion(ctx.Name))
				Expect(err).To(HaveOccurred())
				Expect(errors.IsNotFound(err)).To(BeTrue())
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
						CreationTimestamp: metav1.Now(),
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
			ref := CreateTestRef()
			mutatorErr := openshift.DeploymentConfigMutator(template.NewDefaultEngine())(ctx, &ref)
			Expect(mutatorErr).ToNot(HaveOccurred())

			_, mutatedFetchErr := get.DeploymentConfigWithError(ctx.Namespace, ref.Name+"-"+ref.GetNewVersion(ctx.Name))
			Expect(mutatedFetchErr).ToNot(HaveOccurred())

			revertorErr := openshift.DeploymentConfigRevertor(ctx, &ref)
			Expect(revertorErr).ToNot(HaveOccurred())

			_, revertedFetchErr := get.DeploymentConfigWithError(ctx.Namespace, ref.Name+"-"+ref.GetNewVersion(ctx.Name))
			Expect(revertedFetchErr).To(HaveOccurred())
			Expect(errors.IsNotFound(revertedFetchErr)).To(BeTrue())
		})

	})
})
