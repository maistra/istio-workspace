package session_test

import (
	"context"
	"time"

	"github.com/maistra/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/controller/session"
	"github.com/maistra/istio-workspace/pkg/log"
	"istio.io/api/networking/v1alpha3"

	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/api/pkg/kube/apis/networking/v1alpha3"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
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
		objects    []runtime.Object
		controller reconcile.Reconciler
		req        reconcile.Request
		schema     *runtime.Scheme
		c          client.Client
	)
	GetSession := func(c *client.Client) func(namespace, name string) v1alpha1.Session {
		return func(namespace, name string) v1alpha1.Session {
			s := v1alpha1.Session{}
			err := (*c).Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &s)
			Expect(err).ToNot(HaveOccurred())
			return s
		}
	}(&c)
	/*
		GetStatusRef := func(name string, session v1alpha1.Session) *v1alpha1.RefStatus {
			for _, ref := range session.Status.Refs {
				if ref.Name == name {
					return ref
				}
			}
			return nil
		}
	*/
	JustBeforeEach(func() {
		logf.SetLogger(log.CreateClusterAwareLogger())
		schema, _ = v1alpha1.SchemeBuilder.Build()
		_ = corev1.AddToScheme(schema)
		_ = appsv1.AddToScheme(schema)
		_ = istionetwork.AddToScheme(schema)

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
			objects = []runtime.Object{
				&v1alpha1.Session{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-session",
						Namespace: "test",
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
				},
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "ratings-v1",
						Namespace:         "test",
						CreationTimestamp: metav1.Time{Time: time.Now()},
						Labels: map[string]string{
							"app":     "x",
							"version": "v1",
						},
					},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"app":     "x",
									"version": "v1",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:            "ratings",
										Image:           "x:x:x",
										ImagePullPolicy: "Always",
									},
								},
							},
						},
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app":     "x",
								"version": "v1",
							},
						},
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service",
						Namespace: "test",
					},
					Spec: corev1.ServiceSpec{
						Selector: map[string]string{
							"app": "x",
						},
					},
				},
				&istionetwork.VirtualService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-vs",
						Namespace: "test",
					},
					Spec: v1alpha3.VirtualService{
						Http: []*v1alpha3.HTTPRoute{
							{
								Route: []*v1alpha3.HTTPRouteDestination{
									{
										Destination: &v1alpha3.Destination{
											Host:   "test-service",
											Subset: "v1",
										},
									},
								},
							},
						},
					},
				},
				&istionetwork.DestinationRule{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dr",
						Namespace: "test",
					},
					Spec: istiov1alpha3.DestinationRule{
						Host: "test-service",
						Subsets: []*istiov1alpha3.Subset{
							&istiov1alpha3.Subset{
								Name: "v1",
								Labels: map[string]string{
									"version": "v1",
								},
							},
						},
					},
				},
			}
		})

		It("with update", func() {
			res, err := controller.Reconcile(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Requeue).To(BeFalse())

			target := objects[0].(*v1alpha1.Session)
			target.Spec.Refs[0].Args["image"] = "y:y:y"

			res, err = controller.Reconcile(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Requeue).To(BeFalse())

			sess := GetSession("test", "test-session")

			/*
				b, _ := yaml.Marshal(sess)
				fmt.Println(string(b))
			*/

			Expect(sess.Status.Refs).To(HaveLen(1))
			Expect(sess.Status.Refs[0].Resources).To(HaveLen(3))
			Expect(sess.Status.Refs[0].Targets).To(HaveLen(2))
		})
	})
})
