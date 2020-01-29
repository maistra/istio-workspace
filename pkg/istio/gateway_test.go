package istio

import (
	"github.com/maistra/istio-workspace/pkg/model"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/api/pkg/kube/apis/networking/v1alpha3"
)

var _ = Describe("Operations for istio gateway kind", func() {

	Context("manipulation", func() {
		var (
			ctx     model.SessionContext
			gateway istionetwork.Gateway
		)
		BeforeEach(func() {
			ctx = model.SessionContext{
				Name: "gw-test",
				Route: model.Route{
					Type:  "header",
					Name:  "test",
					Value: "x",
				},
			}
		})

		Context("mutators", func() {

			JustBeforeEach(func() {
				gateway = istionetwork.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gateway",
						Namespace: "test",
					},
					Spec: v1alpha3.Gateway{
						Selector: map[string]string{
							"istio": "ingressgateway",
						},
						Servers: []*v1alpha3.Server{
							{
								Port: &v1alpha3.Port{
									Protocol: "HTTP",
									Name:     "http",
									Number:   80,
								},
								Hosts: []string{
									"domain.com",
								},
							},
						},
					},
				}
			})

			It("add gateway", func() {
				gw, err := mutateGateway(ctx, gateway)
				Expect(err).ToNot(HaveOccurred())

				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
			})
		})

		Context("revertors", func() {

			JustBeforeEach(func() {
				gateway = istionetwork.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gateway",
						Namespace: "test",
					},
					Spec: v1alpha3.Gateway{
						Selector: map[string]string{
							"istio": "ingressgateway",
						},
						Servers: []*v1alpha3.Server{
							{
								Port: &v1alpha3.Port{
									Protocol: "HTTP",
									Name:     "http",
									Number:   80,
								},
								Hosts: []string{
									"domain.com",
									"test.domain.com",
								},
							},
						},
					},
				}
			})

			It("remove gateway", func() {
				gw, err := revertGateway(ctx, gateway)
				Expect(err).ToNot(HaveOccurred())

				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(1))
			})
		})
	})
})
