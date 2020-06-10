package istio

import (
	"fmt"

	"github.com/maistra/istio-workspace/pkg/model"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

			It("add single session", func() {
				gw, err := mutateGateway(ctx, gateway)
				Expect(err).ToNot(HaveOccurred())

				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com", "gw-test.domain.com"))
			})
			It("add multipe session", func() {
				gw, err := mutateGateway(ctx, gateway)
				Expect(err).ToNot(HaveOccurred())

				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com", "gw-test.domain.com"))

				ctx2 := model.SessionContext{
					Name: "gw-test2",
					Route: model.Route{
						Type:  "header",
						Name:  "test",
						Value: "x",
					},
				}
				gw, err = mutateGateway(ctx2, gw)
				Expect(err).ToNot(HaveOccurred())

				fmt.Println(gw.Labels[LabelIkeHosts])
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(3))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com", "gw-test.domain.com", "gw-test2.domain.com"))
			})
			It("add multipe refs", func() {
				gw, err := mutateGateway(ctx, gateway)
				Expect(err).ToNot(HaveOccurred())

				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com", "gw-test.domain.com"))

				gw, err = mutateGateway(ctx, gw)
				Expect(err).ToNot(HaveOccurred())

				fmt.Println(gw.Labels[LabelIkeHosts])
				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com", "gw-test.domain.com"))
			})
		})

		Context("revertors", func() {

			JustBeforeEach(func() {
				gateway = istionetwork.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "gateway",
						Namespace:   "test",
						Annotations: map[string]string{LabelIkeHosts: "gw-test.domain.com,gw-test2.domain.com"},
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
									"gw-test.domain.com",
									"gw-test2.domain.com",
								},
							},
						},
					},
				}
			})

			It("single remove", func() {
				gw, err := revertGateway(ctx, gateway)
				Expect(err).ToNot(HaveOccurred())

				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
			})

			It("multiple remove", func() {
				gw, err := revertGateway(ctx, gateway)
				Expect(err).ToNot(HaveOccurred())

				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(2))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com", "gw-test2.domain.com"))

				ctx2 := model.SessionContext{
					Name: "gw-test2",
					Route: model.Route{
						Type:  "header",
						Name:  "test",
						Value: "x",
					},
				}
				gw, err = revertGateway(ctx2, gw)
				Expect(err).ToNot(HaveOccurred())

				Expect(gw.Spec.Servers[0].Hosts).To(HaveLen(1))
				Expect(gw.Spec.Servers[0].Hosts).To(ContainElements("domain.com"))
				Expect(gw.Labels).ToNot(HaveKey(LabelIkeHosts))
			})
		})
	})
})
