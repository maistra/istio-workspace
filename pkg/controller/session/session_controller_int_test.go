package session_test

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/maistra/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/controller/session"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/test/cmd/test-scenario/generator"
	"github.com/maistra/istio-workspace/test/testclient"

	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log" //nolint:depguard //reason registers wrapper as logger
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		logf.SetLogger(log.CreateOperatorAwareLogger("test").WithValues("type", "session_controller_int_test"))

		schema, _ = v1alpha1.SchemeBuilder.Build()
		_ = corev1.AddToScheme(schema)
		_ = appsv1.AddToScheme(schema)
		_ = istionetwork.AddToScheme(schema)

		objs, err := Scenario(schema, namespace, scenario)
		Expect(err).ToNot(HaveOccurred())
		objects = append(objects, objs...)

		c = fake.NewFakeClientWithScheme(schema, objects...)
		get = testclient.New(c)
		controller = session.NewStandaloneReconciler(c, session.DefaultManipulators())
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
			It("it should update the image", func() {
				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-session1",
						Namespace: "test",
					},
				}
				// given - a fresh session
				res, err := controller.Reconcile(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				// when - a ref is updated
				target := get.Session("test", "test-session1")
				target.Spec.Refs[0].Args["image"] = "y:y:y"

				res, err = controller.Reconcile(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Requeue).To(BeFalse())

				// then
				sess := get.Session("test", "test-session1")
				Expect(target.Spec.Refs[0].Args["image"]).To(Equal("y:y:y"))
				Expect(sess.Status.Refs).To(HaveLen(1))
				Expect(sess.Status.Refs[0].Resources).To(HaveLen(5))
				Expect(sess.Status.Refs[0].Targets).To(HaveLen(3))
			})
		})

		Context("when there are multiple sessions", func() {
			It("shared resources should still be in sync on delete", func() {
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
				res1, err := controller.Reconcile(req1)
				Expect(err).ToNot(HaveOccurred())
				Expect(res1.Requeue).To(BeFalse())

				// Given - create second session
				res2, err := controller.Reconcile(req2)
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
				c.Update(context.Background(), &session)

				res1, err = controller.Reconcile(req1)
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
			It("shared resources should be in sync on delete", func() {
				req1 := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-session1",
						Namespace: "test",
					},
				}
				// Given - create first ref
				res1, err := controller.Reconcile(req1)
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
				c.Update(context.Background(), &session)

				// Given - create second ref
				res2, err := controller.Reconcile(req1)
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
				c.Update(context.Background(), &session)

				res1, err = controller.Reconcile(req1)
				Expect(err).ToNot(HaveOccurred())
				Expect(res2.Requeue).To(BeFalse())

				// Then - no vs removed (only gateway connected duplicated)
				vss = get.VirtualServices("test")
				Expect(vss.Items).To(HaveLen(4))

				// Then - same Gateways Hosts still connected, ref01 still need them
				gw = get.Gateway("test", "test-gateway")
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
			})
		})
	})
})

func Scenario(scheme *runtime.Scheme, namespace string, scenarioGenerator func(io.Writer)) ([]runtime.Object, error) {
	generator.Namespace = namespace
	generator.TestImageName = "x:x:x"
	generator.GatewayHost = "test.io"

	buf := new(bytes.Buffer)
	scenarioGenerator(buf)
	filecontent := buf.String()

	objects := []runtime.Object{}

	filechunks := strings.Split(filecontent, "---")
	for _, filechuck := range filechunks {
		if strings.Trim(filechuck, "\n") == "" {
			continue
		}
		decode := serializer.NewCodecFactory(scheme).UniversalDeserializer().Decode
		obj, _, err := decode([]byte(filechuck), nil, nil)
		if err != nil {
			return nil, err
		}
		if robj, ok := obj.(runtime.Object); ok {
			objects = append(objects, robj)
		}
	}
	return objects, nil
}
