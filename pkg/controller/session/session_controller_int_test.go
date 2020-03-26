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

	istionetwork "istio.io/api/pkg/kube/apis/networking/v1alpha3"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
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
		req        reconcile.Request
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
	JustBeforeEach(func() {
		logf.SetLogger(log.CreateOperatorAwareLogger())

		schema, _ = v1alpha1.SchemeBuilder.Build()
		_ = corev1.AddToScheme(schema)
		_ = appsv1.AddToScheme(schema)
		_ = istionetwork.AddToScheme(schema)

		objs, err := Scenario(schema, namespace, scenario)
		Expect(err).ToNot(HaveOccurred())
		objects = append(objects, objs...)

		req = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      "test-session",
				Namespace: "test",
			},
		}
		c = fake.NewFakeClientWithScheme(schema, objects...)
		controller = session.NewStandaloneReconciler(c, session.DefaultManipulators())
	})

	Context("in a complete lifecycle", func() {
		BeforeEach(func() {
			scenario = generator.TestScenario1ThreeServicesInSequence
			objects = append(objects, &v1alpha1.Session{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-session",
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
		})

		It("with update", func() {
			res, err := controller.Reconcile(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Requeue).To(BeFalse())

			target := GetSession("test", "test-session")
			target.Spec.Refs[0].Args["image"] = "y:y:y"

			res, err = controller.Reconcile(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Requeue).To(BeFalse())

			sess := GetSession("test", "test-session")

			Expect(sess.Status.Refs).To(HaveLen(1))
			Expect(sess.Status.Refs[0].Resources).To(HaveLen(3))
			Expect(sess.Status.Refs[0].Targets).To(HaveLen(2))
		})
	})
})

func Scenario(scheme *runtime.Scheme, namespace string, scenarioGenerator func(io.Writer)) ([]runtime.Object, error) {
	generator.Namespace = namespace
	generator.TestImageName = "x:x:x"

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
