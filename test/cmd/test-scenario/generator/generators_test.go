package generator_test

import (
	"github.com/maistra/istio-workspace/test/cmd/test-scenario/generator"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	osappsv1 "github.com/openshift/api/apps/v1"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Operations for test scenario generator", func() {

	var ns = "test"

	Context("basic sub generators", func() {

		validateLivenessProbe := func(template *corev1.PodTemplateSpec) {
			Expect(template.Spec.Containers[0].LivenessProbe).ToNot(BeNil())
		}

		validateReadinessProbe := func(template *corev1.PodTemplateSpec) {
			Expect(template.Spec.Containers[0].ReadinessProbe).ToNot(BeNil())
		}

		When("generating DeploymentConfig", func() {

			It("should be created if entry is correct DeploymentType", func() {
				obj := generator.DeploymentConfig(generator.Entry{Name: "test", DeploymentType: "DeploymentConfig", Namespace: ns})
				Expect(obj).ToNot(BeNil())
			})

			It("should not be created if entry is not correct DeploymentType", func() {
				obj := generator.DeploymentConfig(generator.Entry{Name: "test", DeploymentType: "X", Namespace: ns})
				Expect(obj).To(BeNil())
			})

			It("should create with liveness probe", func() {
				obj := generator.DeploymentConfig(generator.Entry{Name: "test", DeploymentType: "DeploymentConfig", Namespace: ns})
				Expect(obj).To(BeAssignableToTypeOf(&osappsv1.DeploymentConfig{}))
				validateLivenessProbe(obj.(*osappsv1.DeploymentConfig).Spec.Template)
			})

			It("should create with readiness probe", func() {
				obj := generator.DeploymentConfig(generator.Entry{Name: "test", DeploymentType: "DeploymentConfig", Namespace: ns})
				Expect(obj).To(BeAssignableToTypeOf(&osappsv1.DeploymentConfig{}))
				validateReadinessProbe(obj.(*osappsv1.DeploymentConfig).Spec.Template)
			})
		})

		When("generating Gateway", func() {

			BeforeEach(func() {
				generator.Namespace = ""
				generator.GatewayNamespace = ""
			})

			It("should use gateway namespace if defined", func() {
				generator.Namespace = "test"
				generator.GatewayNamespace = "gw-test"

				gateway := generator.Gateway()

				Expect(gateway.(*v1alpha3.Gateway).Namespace).To(Equal("gw-test"))
			})

			It("should use namespace if gateway namespace not defined", func() {
				generator.Namespace = "test"

				gateway := generator.Gateway()

				Expect(gateway.(*v1alpha3.Gateway).Namespace).To(Equal("test"))
			})

		})

		When("generating Deployment", func() {

			It("should be created if entry is correct DeploymentType", func() {
				obj := generator.Deployment(generator.Entry{Name: "test", DeploymentType: "Deployment", Namespace: ns})
				Expect(obj).ToNot(BeNil())
			})

			It("should not be created if entry is not correct DeploymentType", func() {
				obj := generator.Deployment(generator.Entry{Name: "test", DeploymentType: "X", Namespace: ns})
				Expect(obj).To(BeNil())
			})

			It("should create with liveness probe", func() {
				obj := generator.Deployment(generator.Entry{Name: "test", DeploymentType: "Deployment", Namespace: ns})
				Expect(obj).To(BeAssignableToTypeOf(&appsv1.Deployment{}))
				validateLivenessProbe(&obj.(*appsv1.Deployment).Spec.Template)
			})

			It("should create with readiness probe", func() {
				obj := generator.Deployment(generator.Entry{Name: "test", DeploymentType: "Deployment", Namespace: ns})
				Expect(obj).To(BeAssignableToTypeOf(&appsv1.Deployment{}))
				validateReadinessProbe(&obj.(*appsv1.Deployment).Spec.Template)
			})
		})
	})

})
