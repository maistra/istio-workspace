package openshift_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/maistra/istio-workspace/pkg/k8s"
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
			KindName: model.ParseRefKindName("test-ref"),
			Strategy: "telepresence",
			Args:     map[string]string{"version": "0.103"},
		}
	}
	CreateEmptyTestLocatorStore := func() model.LocatorStore {
		return model.LocatorStore{}
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
			ref := model.Ref{KindName: model.ParseRefKindName("test-ref-other")}
			store := CreateEmptyTestLocatorStore()
			openshift.DeploymentConfigLocator(ctx, ref, store.Store, store.Report)
			Expect(store.Store(openshift.DeploymentConfigKind)).To(HaveLen(0))
		})

		It("should report true on found", func() {
			ref := model.Ref{KindName: model.ParseRefKindName("test-ref")}
			store := CreateEmptyTestLocatorStore()
			openshift.DeploymentConfigLocator(ctx, ref, store.Store, store.Report)
			Expect(store.Store(openshift.DeploymentConfigKind)).To(HaveLen(1))
		})

		It("should find with kind", func() {
			ref := model.Ref{KindName: model.ParseRefKindName("deploymentconfig/test-ref")}
			store := CreateEmptyTestLocatorStore()
			openshift.DeploymentConfigLocator(ctx, ref, store.Store, store.Report)
			Expect(store.Store(openshift.DeploymentConfigKind)).To(HaveLen(1))
		})

		It("should find with abbrev kind", func() {
			ref := model.Ref{KindName: model.ParseRefKindName("dc/test-ref")}
			store := CreateEmptyTestLocatorStore()
			openshift.DeploymentConfigLocator(ctx, ref, store.Store, store.Report)
			Expect(store.Store(openshift.DeploymentConfigKind)).To(HaveLen(1))
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

		CreateTestLocatorStoreWithRefToBeCreated := func(kind string) model.LocatorStore {
			l := model.LocatorStore{}
			l.Report(model.LocatorStatus{Resource: model.Resource{Kind: kind, Name: "test-ref", Namespace: "test"}, Labels: map[string]string{"version": "v1"}, Action: model.ActionCreate})

			return l
		}

		It("should add reference to cloned deployment", func() {
			ref := CreateTestRef()
			store := CreateTestLocatorStoreWithRefToBeCreated(openshift.DeploymentConfigKind)
			modificatorStore := model.ModificatorStore{}
			openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

			dc := get.DeploymentConfig(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(reference.Get(&dc)).To(HaveLen(1))
		})

		It("should add suffix to the cloned deploymentconfig", func() {
			ref := CreateTestRef()
			store := CreateTestLocatorStoreWithRefToBeCreated(openshift.DeploymentConfigKind)
			modificatorStore := model.ModificatorStore{}

			openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)
			Expect(modificatorStore.Stored).To(HaveLen(1))
			Expect(modificatorStore.Stored[0].Success).To(BeTrue())

			_ = get.DeploymentConfig(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
		})

		It("should remove liveness probe from cloned deployment", func() {
			ref := CreateTestRef()
			store := CreateTestLocatorStoreWithRefToBeCreated(openshift.DeploymentConfigKind)
			modificatorStore := model.ModificatorStore{}
			openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

			deployment := get.DeploymentConfig(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(deployment.Spec.Template.Spec.Containers[0].LivenessProbe).To(BeNil())
		})

		It("should remove readiness probe from cloned deployment", func() {
			ref := CreateTestRef()
			store := CreateTestLocatorStoreWithRefToBeCreated(openshift.DeploymentConfigKind)
			modificatorStore := model.ModificatorStore{}
			openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

			deployment := get.DeploymentConfig(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(deployment.Spec.Template.Spec.Containers[0].ReadinessProbe).To(BeNil())
		})

		It("should update selector", func() {
			ref := CreateTestRef()
			store := CreateTestLocatorStoreWithRefToBeCreated(openshift.DeploymentConfigKind)
			modificatorStore := model.ModificatorStore{}
			openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

			deployment := get.DeploymentConfig(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(deployment.Spec.Selector["version"]).To(BeEquivalentTo(model.GetSha("v1") + "-test"))
		})

		It("should only mutate if Target is of kind DeploymentConfig", func() {
			notMatchingRef := model.Ref{
				KindName: model.RefKindName{Name: "test-ref", Kind: k8s.DeploymentKind},
			}
			store := CreateTestLocatorStoreWithRefToBeCreated(k8s.DeploymentKind)
			modificatorStore := model.ModificatorStore{}

			openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, notMatchingRef, store.Store, modificatorStore.Report)
			Expect(modificatorStore.Stored).To(HaveLen(0))

			_, err := get.DeploymentConfigWithError(ctx.Namespace, notMatchingRef.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("should recreate cloned DeploymentConfig if deleted externally", func() {
			// given a normal setup
			ref := CreateTestRef()
			store := CreateTestLocatorStoreWithRefToBeCreated(openshift.DeploymentConfigKind)
			modificatorStore := model.ModificatorStore{}

			openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

			deployment := get.DeploymentConfig(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(deployment.Spec.Selector["version"]).To(BeEquivalentTo(model.GetSha("v1") + "-test"))

			// when DeploymentConfig is deleted
			err := c.Delete(ctx, &deployment)
			Expect(err).To(Not(HaveOccurred()))

			_, err = get.DeploymentConfigWithError(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(err).To(HaveOccurred())

			// then it should be recreated on next reconcile
			openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

			deployment = get.DeploymentConfig(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(deployment.Spec.Selector["version"]).To(BeEquivalentTo(model.GetSha("v1") + "-test"))
		})

		Context("telepresence mutation strategy", func() {

			It("should change container to telepresence", func() {
				ref := CreateTestRef()
				store := CreateTestLocatorStoreWithRefToBeCreated(openshift.DeploymentConfigKind)
				modificatorStore := model.ModificatorStore{}

				openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

				deployment := get.DeploymentConfig(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
				Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(ContainSubstring("datawire/telepresence-k8s:"))
			})

			It("should change add required env variables", func() {
				ref := CreateTestRef()
				store := CreateTestLocatorStoreWithRefToBeCreated(openshift.DeploymentConfigKind)
				modificatorStore := model.ModificatorStore{}

				openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

				deployment := get.DeploymentConfig(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
				Expect(deployment.Spec.Template.Spec.Containers[0].Env[0].Name).To(Equal("TELEPRESENCE_CONTAINER_NAMESPACE"))
				Expect(deployment.Spec.Template.Spec.Containers[0].Env[0].ValueFrom).ToNot(BeNil())
			})

		})

		Context("existing mutation strategy", func() {

			It("should not create a clone", func() {
				ref := CreateTestRef()
				ref.Strategy = model.StrategyExisting

				store := CreateTestLocatorStoreWithRefToBeCreated(openshift.DeploymentConfigKind)
				modificatorStore := model.ModificatorStore{}

				openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

				_, err := get.DeploymentConfigWithError(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
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

			// Create
			store := CreateEmptyTestLocatorStore()
			modificatorStore := model.ModificatorStore{}
			openshift.DeploymentConfigLocator(ctx, ref, store.Store, store.Report)

			openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)
			Expect(modificatorStore.Stored).To(HaveLen(1))
			Expect(modificatorStore.Stored[0].Error).ToNot(HaveOccurred())

			_, mutatedFetchErr := get.DeploymentConfigWithError(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(mutatedFetchErr).ToNot(HaveOccurred())

			// Setup deleted ref
			ref.Remove = true

			// Revert
			store = CreateEmptyTestLocatorStore()
			modificatorStore = model.ModificatorStore{}
			openshift.DeploymentConfigLocator(ctx, ref, store.Store, store.Report)

			openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)
			Expect(modificatorStore.Stored).To(HaveLen(1))
			Expect(modificatorStore.Stored[0].Error).ToNot(HaveOccurred())

			_, revertedFetchErr := get.DeploymentConfigWithError(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(revertedFetchErr).To(HaveOccurred())
			Expect(errors.IsNotFound(revertedFetchErr)).To(BeTrue())
		})

		It("should be able to detect change in ref", func() {
			ref := CreateTestRef()

			// Create
			store := CreateEmptyTestLocatorStore()
			modificatorStore := model.ModificatorStore{}
			openshift.DeploymentConfigLocator(ctx, ref, store.Store, store.Report)

			openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)
			Expect(modificatorStore.Stored).To(HaveLen(1))
			Expect(modificatorStore.Stored[0].Error).ToNot(HaveOccurred())

			_, mutatedFetchErr := get.DeploymentConfigWithError(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(mutatedFetchErr).ToNot(HaveOccurred())

			// Setup deleted ref
			ref.Remove = true

			// Revert
			store = CreateEmptyTestLocatorStore()
			modificatorStore = model.ModificatorStore{}
			openshift.DeploymentConfigLocator(ctx, ref, store.Store, store.Report)

			openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)
			Expect(modificatorStore.Stored).To(HaveLen(1))
			Expect(modificatorStore.Stored[0].Error).ToNot(HaveOccurred())

			_, revertedFetchErr := get.DeploymentConfigWithError(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(revertedFetchErr).To(HaveOccurred())
			Expect(errors.IsNotFound(revertedFetchErr)).To(BeTrue())
		})

		It("should be able to detect change in ref", func() {
			ref := CreateTestRef()

			// Create
			store := CreateEmptyTestLocatorStore()
			modificatorStore := model.ModificatorStore{}
			openshift.DeploymentConfigLocator(ctx, ref, store.Store, store.Report)

			openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)
			Expect(modificatorStore.Stored).To(HaveLen(1))
			Expect(modificatorStore.Stored[0].Error).ToNot(HaveOccurred())

			_, mutatedFetchErr := get.DeploymentConfigWithError(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(mutatedFetchErr).ToNot(HaveOccurred())

			// Setup deleted ref
			imageName := "docker.io/maistra:latest"
			ref.Strategy = "prepared-image"
			ref.Args = map[string]string{}
			ref.Args["image"] = imageName

			// Revert
			store = CreateEmptyTestLocatorStore()
			modificatorStore = model.ModificatorStore{}
			openshift.DeploymentConfigLocator(ctx, ref, store.Store, store.Report)

			openshift.DeploymentConfigModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)
			Expect(modificatorStore.Stored).To(HaveLen(2))
			Expect(modificatorStore.Stored[0].Error).ToNot(HaveOccurred())

			deployment, mutatedFetchErr := get.DeploymentConfigWithError(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(mutatedFetchErr).ToNot(HaveOccurred())

			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal(imageName))
		})
	})

})
