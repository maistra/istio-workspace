package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	osappsv1 "github.com/openshift/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/maistra/istio-workspace/test/cmd/test-scenario/generator"
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
		Context("deploymentconfig", func() {
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
		Context("deployment", func() {
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
