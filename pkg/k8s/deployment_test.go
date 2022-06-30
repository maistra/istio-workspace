package k8s_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
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

var _ = Describe("Operations for k8s Deployment kind", func() {

	var (
		objects []runtime.Object
		c       client.Client
		ctx     model.SessionContext
		get     *testclient.Getters
	)

	CreateTestRef := func(name string) model.Ref {
		return model.Ref{
			KindName:  model.RefKindName{Name: name},
			Namespace: "test",
			Strategy:  "telepresence",
			Args:      map[string]string{"version": "0.103"},
		}
	}
	CreateEmptyTestLocatorStore := func() model.LocatorStore {
		return model.LocatorStore{}
	}
	CreateTestLocatorStoreWithRefToBeCreated := func(kind string) model.LocatorStore {
		l := model.LocatorStore{}
		l.Report(model.LocatorStatus{Resource: model.Resource{Kind: kind, Name: "test-ref", Namespace: "test"}, Labels: map[string]string{"version": "v1"}, Action: model.ActionCreate})

		return l
	}

	JustBeforeEach(func() {
		schema := runtime.NewScheme()
		err := appsv1.AddToScheme(schema)
		Expect(err).ToNot(HaveOccurred())
		c = fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(objects...).Build()
		get = testclient.New(c)
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
			ref := CreateTestRef("non-existing-ref")
			store := CreateEmptyTestLocatorStore()
			k8s.DeploymentLocator(ctx, ref, store.Store, store.Report)
			Expect(store.Store(k8s.DeploymentKind)).To(HaveLen(0))
		})

		It("should report true on found", func() {
			ref := CreateTestRef("test-ref")
			store := CreateEmptyTestLocatorStore()
			k8s.DeploymentLocator(ctx, ref, store.Store, store.Report)
			Expect(store.Store(k8s.DeploymentKind)).To(HaveLen(1))
		})

		It("should find with kind", func() {
			ref := model.Ref{Namespace: "test", KindName: model.ParseRefKindName("deployment/test-ref")}
			store := CreateEmptyTestLocatorStore()
			k8s.DeploymentLocator(ctx, ref, store.Store, store.Report)
			Expect(store.Store(k8s.DeploymentKind)).To(HaveLen(1))
		})

	})

	Context("modificators", func() {

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
											ProbeHandler: v1.ProbeHandler{
												HTTPGet: &v1.HTTPGetAction{
													Path: "/healthz",
													Port: intstr.FromInt(9080),
												},
											},
										},
										ReadinessProbe: &v1.Probe{
											ProbeHandler: v1.ProbeHandler{
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
			ref := CreateTestRef("test-ref")
			store := CreateTestLocatorStoreWithRefToBeCreated(k8s.DeploymentKind)
			modificatorStore := model.ModificatorStore{}
			k8s.DeploymentModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

			d := get.Deployment(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(reference.Get(&d)).To(HaveLen(1))
		})

		It("should add suffix to the cloned deployment", func() {
			ref := CreateTestRef("test-ref")
			store := CreateTestLocatorStoreWithRefToBeCreated(k8s.DeploymentKind)
			modificatorStore := model.ModificatorStore{}
			k8s.DeploymentModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)
			Expect(modificatorStore.Stored).To(HaveLen(1))
			Expect(modificatorStore.Stored[0].Success).To(BeTrue())

			_ = get.Deployment(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
		})

		It("should remove liveness probe from cloned deployment", func() {
			ref := CreateTestRef("test-ref")
			store := CreateTestLocatorStoreWithRefToBeCreated(k8s.DeploymentKind)
			modificatorStore := model.ModificatorStore{}
			k8s.DeploymentModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

			deployment := get.Deployment(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(deployment.Spec.Template.Spec.Containers[0].LivenessProbe).To(BeNil())
		})

		It("should remove readiness probe from cloned deployment", func() {
			ref := CreateTestRef("test-ref")
			store := CreateTestLocatorStoreWithRefToBeCreated(k8s.DeploymentKind)
			modificatorStore := model.ModificatorStore{}
			k8s.DeploymentModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

			deployment := get.Deployment(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(deployment.Spec.Template.Spec.Containers[0].ReadinessProbe).To(BeNil())
		})

		It("should update selector", func() {
			ref := CreateTestRef("test-ref")
			store := CreateTestLocatorStoreWithRefToBeCreated(k8s.DeploymentKind)
			modificatorStore := model.ModificatorStore{}
			k8s.DeploymentModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

			deployment := get.Deployment(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(deployment.Spec.Template.Spec.Containers[0].ReadinessProbe).To(BeNil())
			Expect(deployment.Spec.Selector.MatchLabels["version"]).To(BeEquivalentTo(model.GetSha("v1") + "-test"))
		})

		// different action in the store
		It("should only mutate if Target is of kind Deployment", func() {
			notMatchingRef := model.Ref{
				KindName: model.RefKindName{Name: "test-ref", Kind: openshift.DeploymentConfigKind},
			}
			store := CreateTestLocatorStoreWithRefToBeCreated(openshift.DeploymentConfigKind)
			modificatorStore := model.ModificatorStore{}
			k8s.DeploymentModificator(template.NewDefaultEngine())(ctx, notMatchingRef, store.Store, modificatorStore.Report)
			Expect(modificatorStore.Stored).To(HaveLen(0))

			_, err := get.DeploymentWithError(ctx.Namespace, notMatchingRef.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("should recreate cloned Deployment if deleted externally", func() {
			// given a normal setup
			ref := CreateTestRef("test-ref")
			store := CreateTestLocatorStoreWithRefToBeCreated(k8s.DeploymentKind)
			modificatorStore := model.ModificatorStore{}

			k8s.DeploymentModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

			deployment := get.Deployment(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(deployment.Spec.Selector.MatchLabels["version"]).To(BeEquivalentTo(model.GetSha("v1") + "-test"))

			// when Deployment is deleted
			err := c.Delete(ctx, &deployment)
			Expect(err).To(Not(HaveOccurred()))

			_, err = get.DeploymentWithError(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(err).To(HaveOccurred())

			// then it should be recreated on next reconcile
			k8s.DeploymentModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

			deployment = get.Deployment(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
			Expect(deployment.Spec.Selector.MatchLabels["version"]).To(BeEquivalentTo(model.GetSha("v1") + "-test"))
		})

		Context("telepresence mutation strategy", func() {

			It("should change container to telepresence", func() {
				ref := CreateTestRef("test-ref")
				store := CreateTestLocatorStoreWithRefToBeCreated(k8s.DeploymentKind)
				modificatorStore := model.ModificatorStore{}
				k8s.DeploymentModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

				deployment := get.Deployment(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
				Expect(deployment.Spec.Template.Spec.Containers[0].ReadinessProbe).To(BeNil())
				Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(ContainSubstring("datawire/telepresence-k8s:"))
			})

			It("should change add required env variables", func() {
				ref := CreateTestRef("test-ref")
				store := CreateTestLocatorStoreWithRefToBeCreated(k8s.DeploymentKind)
				modificatorStore := model.ModificatorStore{}
				k8s.DeploymentModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

				deployment := get.Deployment(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
				Expect(deployment.Spec.Template.Spec.Containers[0].Env[0].Name).To(Equal("TELEPRESENCE_CONTAINER_NAMESPACE"))
				Expect(deployment.Spec.Template.Spec.Containers[0].Env[0].ValueFrom).ToNot(BeNil())
			})

		})

		Context("existing mutation strategy", func() {

			It("should not create a clone", func() {
				ref := CreateTestRef("test-ref")
				ref.Strategy = model.StrategyExisting

				store := CreateTestLocatorStoreWithRefToBeCreated(k8s.DeploymentKind)
				modificatorStore := model.ModificatorStore{}
				k8s.DeploymentModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)

				_, err := get.DeploymentWithError(ctx.Namespace, ref.KindName.Name+"-"+model.GetCreatedVersion(store.Store, ctx.Name))
				Expect(err).To(HaveOccurred())
				Expect(errors.IsNotFound(err)).To(BeTrue())
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

		It("should update strategy when ref change", func() {
			ref := CreateTestRef("test-ref")

			// Create
			store := CreateEmptyTestLocatorStore()
			modificatorStore := model.ModificatorStore{}
			k8s.DeploymentLocator(ctx, ref, store.Store, store.Report)

			k8s.DeploymentModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)
			Expect(modificatorStore.Stored).To(HaveLen(1))
			Expect(modificatorStore.Stored[0].Error).ToNot(HaveOccurred())

			newName := ref.KindName.Name + "-" + model.GetCreatedVersion(store.Store, ctx.Name)
			_, mutatedFetchErr := get.DeploymentWithError(ctx.Namespace, newName)
			Expect(mutatedFetchErr).ToNot(HaveOccurred())

			// Setup deleted ref
			imageName := "docker.io/maistra:latest"
			ref.Strategy = "prepared-image"
			ref.Args = map[string]string{}
			ref.Args["image"] = imageName

			// Revert
			store = CreateEmptyTestLocatorStore()
			modificatorStore = model.ModificatorStore{}
			k8s.DeploymentLocator(ctx, ref, store.Store, store.Report)

			k8s.DeploymentModificator(template.NewDefaultEngine())(ctx, ref, store.Store, modificatorStore.Report)
			Expect(modificatorStore.Stored).To(HaveLen(2))
			Expect(modificatorStore.Stored[0].Error).ToNot(HaveOccurred())

			deployment, mutatedFetchErr := get.DeploymentWithError(ctx.Namespace, newName)
			Expect(mutatedFetchErr).ToNot(HaveOccurred())

			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal(imageName))
		})

	})
})
