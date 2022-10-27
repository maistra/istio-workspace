package session_test

import (
	"github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/controllers/session"
	"github.com/maistra/istio-workspace/pkg/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Basic model conversion", func() {

	var sess v1alpha1.Session

	Context("route to route", func() {
		var (
			route model.Route
		)
		JustBeforeEach(func() {
			route = session.ConvertAPIRouteToModelRoute(&sess)
		})
		Context("missing", func() {
			BeforeEach(func() {
				sess = v1alpha1.Session{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-session",
					},
					Spec: v1alpha1.SessionSpec{
						Route: v1alpha1.Route{},
					},
					Status: v1alpha1.SessionStatus{},
				}
			})

			It("should default if no route defined", func() {
				Expect(route.Type).To(Equal(session.RouteStrategyHeader))
				Expect(route.Name).To(Equal(session.DefaultRouteHeaderName))
				Expect(route.Value).To(Equal(sess.ObjectMeta.Name))
			})
		})
		Context("exists", func() {
			BeforeEach(func() {
				sess = v1alpha1.Session{
					Spec: v1alpha1.SessionSpec{
						Route: v1alpha1.Route{
							Type:  "header",
							Name:  "x",
							Value: "y",
						},
					},
					Status: v1alpha1.SessionStatus{},
				}
			})

			It("should map route if provided", func() {
				Expect(route.Type).To(Equal(sess.Spec.Route.Type))
				Expect(route.Name).To(Equal(sess.Spec.Route.Name))
				Expect(route.Value).To(Equal(sess.Spec.Route.Value))
			})
		})

	})
})
