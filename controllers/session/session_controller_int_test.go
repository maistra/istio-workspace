package session_test

import (
	"bytes"
	"context"
	"io"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	osappsv1 "github.com/openshift/api/apps/v1"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/controllers/session"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/model"
	"github.com/maistra/istio-workspace/pkg/template"
	"github.com/maistra/istio-workspace/test"
	"github.com/maistra/istio-workspace/test/cmd/test-scenario/generator"
	"github.com/maistra/istio-workspace/test/testclient"
)

var _ = Describe("Complete session manipulation", func() {
	var (
		namespace  = "test"
		objects    []runtime.Object
		controller reconcile.Reconciler
		schema     *runtime.Scheme
		c          client.Client
		scenario   func(io.Writer)
		get        *testclient.Getters
	)

	JustBeforeEach(func() {
		log.SetLogger(log.CreateOperatorAwareLogger("test").WithValues("type", "session_controller_int_test"))

		schema, _ = v1alpha1.SchemeBuilder.Build()
		_ = corev1.AddToScheme(schema)
		_ = appsv1.AddToScheme(schema)
		_ = istionetwork.AddToScheme(schema)
		_ = osappsv1.Install(schema)

		objs, err := Scenario(schema, namespace, scenario)
		Expect(err).ToNot(HaveOccurred())
		objects = append(objects, objs...)

		c = fake.NewClientBuilder().WithScheme(schema).WithRuntimeObjects(objects...).Build()
		get = testclient.New(c)
		controller = session.NewStandaloneReconciler(c, session.DefaultManipulators(), session.DefaultValidators()...)
	})

	Context("in a complete lifecycle", func() {
		BeforeEach(func() {
			scenario = generator.TestScenario1HTTPThreeServicesInSequence
			objects = []runtime.Object{}
			objects = append(objects, &v1alpha1.Session{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-session1",
					Namespace: namespace,
				},
				Spec: v1alpha1.SessionSpec{
					Refs: []v1alpha1.Ref{
						{
							Name:     "ratings-v1",
							Strategy: "prepared-image",
							Args: map[string]string{
								"image": "x:x:x",
							},
						},
					},
				},
			})
			objects = append(objects, &v1alpha1.Session{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-session2",
					Namespace: namespace,
				},
				Spec: v1alpha1.SessionSpec{
					Refs: []v1alpha1.Ref{
						{
							Name:     "reviews-v1",
							Strategy: "prepared-image",
							Args: map[string]string{
								"image": "x:x:x",
							},
						},
					},
				},
			})
		})

		Context("when a ref is updated", func() {

			It("should update the image", func() {
				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-session1",
						Namespace: "test",
					},
				}
				// given - a fresh session
				res, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				// when - a ref is updated
				target := get.Session("test", "test-session1")
				target.Spec.Refs[0].Args["image"] = "y:y:y"
				c.Update(context.Background(), &target)

				res, err = controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				// then
				sess := get.Session("test", "test-session1")
				Expect(sess.Spec.Refs[0].Args["image"]).To(Equal("y:y:y"))

				deployment := get.Deployment("test", "ratings-v1-"+model.GetSha("v1")+"-test-session1")
				Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal("y:y:y"))
			})
		})

		Context("when there are multiple sessions", func() {

			It("should sync resources on delete", func() {
				req1 := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-session1",
						Namespace: "test",
					},
				}
				req2 := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-session2",
						Namespace: "test",
					},
				}

				// Given - create first session
				res1, err := controller.Reconcile(context.Background(), req1)
				Expect(err).ToNot(HaveOccurred())
				Expect(res1.Requeue).To(BeFalse())

				// Given - create second session
				res2, err := controller.Reconcile(context.Background(), req2)
				Expect(err).ToNot(HaveOccurred())
				Expect(res2.Requeue).To(BeFalse())

				// Given - sane creation
				gw := get.Gateway("test", "test-gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(3))

				vss := get.VirtualServices("test")
				Expect(vss.Items).To(HaveLen(5))

				// When - delete first session
				session := get.Session("test", "test-session1")
				now := metav1.Now()
				session.DeletionTimestamp = &now

				updateErr := c.Update(context.Background(), &session)
				Expect(updateErr).ToNot(HaveOccurred())

				res1, err = controller.Reconcile(context.Background(), req1)
				Expect(err).ToNot(HaveOccurred())
				Expect(res1.Requeue).To(BeFalse())

				// Then - virtualservice was cleaned up
				vss = get.VirtualServices("test")
				Expect(vss.Items).To(HaveLen(4))

				// Then - gateway was cleaned up
				gw = get.Gateway("test", "test-gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
			})
		})

		Context("when there are multiple refs in a session", func() {
			It("should sync shared resources on delete", func() {
				req1 := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-session1",
						Namespace: "test",
					},
				}
				// Given - create first ref
				res1, err := controller.Reconcile(context.Background(), req1)
				Expect(err).ToNot(HaveOccurred())
				Expect(res1.Requeue).To(BeFalse())

				session := get.Session("test", "test-session1")

				session.Spec.Refs = append(session.Spec.Refs,
					v1alpha1.Ref{
						Name:     "reviews-v1",
						Strategy: "prepared-image",
						Args: map[string]string{
							"image": "x:x:x",
						},
					},
				)
				updateErr := c.Update(context.Background(), &session)
				Expect(updateErr).ToNot(HaveOccurred())

				// Given - create second ref
				res2, err := controller.Reconcile(context.Background(), req1)
				Expect(err).ToNot(HaveOccurred())
				Expect(res2.Requeue).To(BeFalse())

				// Given - sane creation
				gw := get.Gateway("test", "test-gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))

				vss := get.VirtualServices("test")
				Expect(vss.Items).To(HaveLen(4))

				// When - delete first ref
				session = get.Session("test", "test-session1")
				session.Spec.Refs = []v1alpha1.Ref{session.Spec.Refs[1]}

				updateErr = c.Update(context.Background(), &session)
				Expect(updateErr).ToNot(HaveOccurred())

				res1, err = controller.Reconcile(context.Background(), req1)
				Expect(err).ToNot(HaveOccurred())
				Expect(res2.Requeue).To(BeFalse())

				// Then - same Gateways Hosts still connected, ref01 still need them
				gw = get.Gateway("test", "test-gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))

				// Then - no vs removed (only gateway connected duplicated)
				vss = get.VirtualServices("test")
				Expect(vss.Items).To(HaveLen(4))

			})

			It("should mutate all refs added to Session", func() {
				req1 := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-session1",
						Namespace: "test",
					},
				}
				// Given - create first ref
				res1, err := controller.Reconcile(context.Background(), req1)
				Expect(err).ToNot(HaveOccurred())
				Expect(res1.Requeue).To(BeFalse())

				session := get.Session("test", "test-session1")

				session.Spec.Refs = append(session.Spec.Refs,
					v1alpha1.Ref{
						Name:     "reviews-v1",
						Strategy: "prepared-image",
						Args: map[string]string{
							"image": "x:x:x",
						},
					},
				)
				updateErr := c.Update(context.Background(), &session)
				Expect(updateErr).ToNot(HaveOccurred())

				// Given - create second ref
				res2, err := controller.Reconcile(context.Background(), req1)
				Expect(err).ToNot(HaveOccurred())
				Expect(res2.Requeue).To(BeFalse())

				// Then - all mutations should be successful
				session = get.Session("test", "test-session1")

				Expect(*session.Status.State).To(Equal(v1alpha1.StateSuccess))
			})
		})

		Context("when calculating components readiness", func() {
			It("should all show ready on success", func() {
				req1 := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-session1",
						Namespace: "test",
					},
				}
				// Given - create first ref
				res1, err := controller.Reconcile(context.Background(), req1)
				Expect(err).ToNot(HaveOccurred())
				Expect(res1.Requeue).To(BeFalse())

				// Then - all components should be success
				session := get.Session("test", "test-session1")

				Expect(*session.Status.State).To(Equal(v1alpha1.StateSuccess))
				Expect(session.Status.Readiness.Components.Pending).To((HaveLen(0)))
				Expect(session.Status.Readiness.Components.UnReady).To((HaveLen(0)))
				Expect(session.Status.Readiness.Components.Ready).ToNot((HaveLen(0)))
			})

			It("should show pending before execution and missing modification, and unready on unsuccess", func() {
				testLocator := func(ctx model.SessionContext, ref model.Ref, store model.LocatorStatusStore, report model.LocatorStatusReporter) error {
					// When - three Resources are found
					report(model.LocatorStatus{
						Resource: model.Resource{
							Namespace: ctx.Namespace,
							Kind:      "Pod",
							Name:      "TestPod",
						},
						Action: model.ActionLocated,
					})
					report(model.LocatorStatus{
						Resource: model.Resource{
							Namespace: ctx.Namespace,
							Kind:      "Pod",
							Name:      "TestPod2",
						},
						Action: model.ActionLocated,
					})
					report(model.LocatorStatus{
						Resource: model.Resource{
							Namespace: ctx.Namespace,
							Kind:      "Pod",
							Name:      "TestPod3",
						},
						Action: model.ActionLocated,
					})

					return nil
				}
				testRegistrar := func() (client.Object, model.Modificator) {
					return &corev1.Pod{}, func(context model.SessionContext, ref model.Ref, store model.LocatorStatusStore, reporter model.ModificatorStatusReporter) {

						// Then - Verify all are in pending stage before modification stage
						session := get.Session("test", "test-session1")
						Expect(session.Status.Readiness.Components.Pending).To((HaveLen(3)))

						// When - One is reported unsuccess
						reporter(model.ModificatorStatus{
							LocatorStatus: store("Pod")[0],
							Success:       false,
						})
						// When - One is reported success
						reporter(model.ModificatorStatus{
							LocatorStatus: store("Pod")[1],
							Success:       true,
						})
					}
				}
				testManipulator := session.Manipulators{Locators: []model.Locator{testLocator}, Handlers: []model.ModificatorRegistrar{testRegistrar}}
				controller = session.NewStandaloneReconciler(c, testManipulator)

				req1 := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-session1",
						Namespace: "test",
					},
				}
				// Given - create first ref
				res1, err := controller.Reconcile(context.Background(), req1)
				Expect(err).ToNot(HaveOccurred())
				Expect(res1.Requeue).To(BeFalse())

				session := get.Session("test", "test-session1")

				// Then - state should be Failed
				Expect(*session.Status.State).To(Equal(v1alpha1.StateFailed))
				// Then - one should be pending
				Expect(session.Status.Readiness.Components.Pending).To((HaveLen(1)))
				// Then - one should be unready
				Expect(session.Status.Readiness.Components.UnReady).To((HaveLen(1)))
				// Then - one should be ready
				Expect(session.Status.Readiness.Components.Ready).To((HaveLen(1)))
			})
		})
	})
	Context("when validation rules are triggered", func() {
		BeforeEach(func() {
			objects = []runtime.Object{}
			objects = append(objects, &v1alpha1.Session{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-session1",
					Namespace: namespace,
				},
				Spec: v1alpha1.SessionSpec{
					Refs: []v1alpha1.Ref{
						{
							Name:     "ratings-v1",
							Strategy: "prepared-image",
							Args: map[string]string{
								"image": "x:x:x",
							},
						},
					},
				},
			})
			objects = append(objects, &v1alpha1.Session{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-session2",
					Namespace: namespace,
				},
				Spec: v1alpha1.SessionSpec{
					Refs: []v1alpha1.Ref{
						{
							Name:     "reviews-v1",
							Strategy: "prepared-image",
							Args: map[string]string{
								"image": "x:x:x",
							},
						},
					},
				},
			})
		})

		getCondition := func(session v1alpha1.Session, t string) *v1alpha1.Condition {
			for _, con := range session.Status.Conditions {
				if *con.Type == t {
					return con
				}
			}

			return nil
		}
		Context("with missing target", func() {
			BeforeEach(func() {
				scenario = generator.TestScenario1HTTPThreeServicesInSequence
			})
			It("should fail on missing deployment", func() {
				res := appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ratings-v1",
						Namespace: "test",
					},
				}
				c.Delete(context.Background(), &res)

				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-session1",
						Namespace: "test",
					},
				}

				// Given - create first ref
				res1, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res1.Requeue).To(BeFalse())

				session := get.Session("test", "test-session1")
				Expect(*session.Status.State).To(Equal(v1alpha1.StateFailed))

				Expect(*getCondition(session, "FindDestinationRule").Status).To(Equal("false"))
				Expect(*getCondition(session, "FindVirtualService").Status).To(Equal("false"))
				Expect(*getCondition(session, "FindTarget").Status).To(Equal("false"))
			})
		})

		Context("with missing destination rule", func() {
			BeforeEach(func() {
				scenario = generator.IncompleteMissingDestinationRules
			})
			It("should fail on missing deployment", func() {
				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-session1",
						Namespace: "test",
					},
				}

				// Given - create first ref
				res1, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res1.Requeue).To(BeFalse())

				session := get.Session("test", "test-session1")
				Expect(*session.Status.State).To(Equal(v1alpha1.StateFailed))

				Expect(*getCondition(session, "FindDestinationRule").Status).To(Equal("false"))
				Expect(*getCondition(session, "FindVirtualService").Status).To(Equal("true"))
				Expect(*getCondition(session, "FindTarget").Status).To(Equal("true"))
			})
		})

		Context("with missing virtual service", func() {
			BeforeEach(func() {
				scenario = generator.IncompleteMissingVirtualServices
			})
			It("should fail on missing deployment", func() {
				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-session1",
						Namespace: "test",
					},
				}

				// Given - create first ref
				res1, err := controller.Reconcile(context.Background(), req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res1.Requeue).To(BeFalse())

				session := get.Session("test", "test-session1")
				Expect(*session.Status.State).To(Equal(v1alpha1.StateFailed))

				Expect(*getCondition(session, "FindDestinationRule").Status).To(Equal("true"))
				Expect(*getCondition(session, "FindVirtualService").Status).To(Equal("false"))
				Expect(*getCondition(session, "FindTarget").Status).To(Equal("true"))
			})
		})

	})
	Context("with dynamically loaded templates", func() {
		var restoreEnvVars func()

		BeforeEach(func() {
			scenario = generator.TestScenario1HTTPThreeServicesInSequence
			objects = []runtime.Object{}
			objects = append(objects, &v1alpha1.Session{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-session1",
					Namespace: namespace,
				},
				Spec: v1alpha1.SessionSpec{
					Refs: []v1alpha1.Ref{
						{
							Name:     "ratings-v1",
							Strategy: "telepresence",
						},
					},
				},
			})

			tmpDir := test.TmpDir(GinkgoT(), "template")
			test.TmpFile(GinkgoT(), tmpDir+"/telepresence.tpl", `
[
	{"op": "replace", "path": "/metadata/name", "value": "{{.Data.Value "/metadata/name"}}-custom-template"},
	{"op": "remove", "path": "/metadata/resourceVersion"}
]
`)
			restoreEnvVars = test.TemporaryEnvVars(template.TemplatePath, tmpDir)
		})

		AfterEach(func() {
			restoreEnvVars()
			test.CleanUpTmpFiles(GinkgoT())
		})

		It("should ensure template was called", func() {
			req1 := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-session1",
					Namespace: "test",
				},
			}
			// Given - create first ref
			res1, err := controller.Reconcile(context.Background(), req1)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1.Requeue).To(BeFalse())

			_, err = get.DeploymentWithError("test", "ratings-v1-custom-template")
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

func Scenario(scheme *runtime.Scheme, namespace string, scenarioGenerator func(io.Writer)) ([]runtime.Object, error) {
	generator.Namespace = namespace
	generator.TestImageName = "x:x:x"
	generator.GatewayHost = "test.io"

	buf := new(bytes.Buffer)
	scenarioGenerator(buf)
	fileContent := buf.String()

	objects := []runtime.Object{}

	fileChunks := strings.Split(fileContent, "---")
	for _, fileChunk := range fileChunks {
		if strings.Trim(fileChunk, "\n") == "" {
			continue
		}
		decode := serializer.NewCodecFactory(scheme).UniversalDeserializer().Decode
		obj, _, err := decode([]byte(fileChunk), nil, nil)
		if err != nil {
			return nil, err
		}
		objects = append(objects, obj)
	}

	return objects, nil
}
