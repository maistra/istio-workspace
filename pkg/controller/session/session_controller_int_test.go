package session_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/maistra/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/controller/session"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/test/cmd/test-scenario/generator"

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
		err        error
		controller reconcile.Reconciler
		schema     *runtime.Scheme
		c          client.Client
		scenario   func(io.Writer)
	)
	GetSession := func(c *client.Client) func(namespace, name string) v1alpha1.Session {
		return func(namespace, name string) v1alpha1.Session {
			s := v1alpha1.Session{}
			err = (*c).Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &s)
			Expect(err).ToNot(HaveOccurred())
			return s
		}
	}(&c)
	GetGateway := func(c *client.Client) func(namespace, name string) istionetwork.Gateway {
		return func(namespace, name string) istionetwork.Gateway {
			s := istionetwork.Gateway{}
			err = (*c).Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &s)
			Expect(err).ToNot(HaveOccurred())
			return s
		}
	}(&c)
	GetVirtualServices := func(c *client.Client) func(namespace string) istionetwork.VirtualServiceList {
		return func(namespace string) istionetwork.VirtualServiceList {
			s := istionetwork.VirtualServiceList{}
			err = (*c).List(context.Background(), &s, client.InNamespace(namespace))
			Expect(err).ToNot(HaveOccurred())
			return s
		}
	}(&c)
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

		It("with ref update", func() {
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-session1",
					Namespace: "test",
				},
			}

			res, err := controller.Reconcile(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Requeue).To(BeFalse())

			target := GetSession("test", "test-session1")
			target.Spec.Refs[0].Args["image"] = "y:y:y"

			res, err = controller.Reconcile(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Requeue).To(BeFalse())

			sess := GetSession("test", "test-session1")
			Expect(target.Spec.Refs[0].Args["image"]).To(Equal("y:y:y"))
			Expect(sess.Status.Refs).To(HaveLen(1))
			Expect(sess.Status.Refs[0].Resources).To(HaveLen(5))
			Expect(sess.Status.Refs[0].Targets).To(HaveLen(3))
		})

		It("with multiple sessions", func() {
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

			// Create first session
			res1, err := controller.Reconcile(req1)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1.Requeue).To(BeFalse())

			// Create second session
			res2, err := controller.Reconcile(req2)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2.Requeue).To(BeFalse())

			// Verify
			gw := GetGateway("test", "test-gateway")
			Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(3))

			vss := GetVirtualServices("test")
			Expect(vss.Items).To(HaveLen(5))

			// Delete first session
			session := GetSession("test", "test-session1")
			now := metav1.Now()
			session.DeletionTimestamp = &now
			c.Update(context.Background(), &session)

			res1, err = controller.Reconcile(req1)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1.Requeue).To(BeFalse())

			// Verify
			vss = GetVirtualServices("test")
			Expect(vss.Items).To(HaveLen(4))

		})

		FIt("with multiple ref", func() {
			req1 := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-session1",
					Namespace: "test",
				},
			}
			// Create first ref
			res1, err := controller.Reconcile(req1)
			Expect(err).ToNot(HaveOccurred())
			Expect(res1.Requeue).To(BeFalse())

			session := GetSession("test", "test-session1")

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

			// Create second ref
			res2, err := controller.Reconcile(req1)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2.Requeue).To(BeFalse())

			// Verify
			gw := GetGateway("test", "test-gateway")
			fmt.Println(gw.Spec.Servers[0].Hosts)
			Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))

			vss := GetVirtualServices("test")
			Expect(vss.Items).To(HaveLen(4))

			// Delete first ref
			session = GetSession("test", "test-session1")
			session.Spec.Refs = []v1alpha1.Ref{session.Spec.Refs[1]}
			c.Update(context.Background(), &session)

			res1, err = controller.Reconcile(req1)
			Expect(err).ToNot(HaveOccurred())
			Expect(res2.Requeue).To(BeFalse())

			// Verify - no vs removed (only gateway connected duplicated)
			vss = GetVirtualServices("test")
			for _, vs := range vss.Items {
				fmt.Println(vs.Name)
			}
			Expect(vss.Items).To(HaveLen(4))

			// Verify - same Gateways Hosts still connected, ref01 still need them
			gw = GetGateway("test", "test-gateway")
			fmt.Println(gw.Spec.Servers[0].Hosts)
			Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
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
