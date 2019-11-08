package parser_test

import (
	"github.com/maistra/istio-workspace/test"
	"github.com/maistra/istio-workspace/version"

	. "github.com/maistra/istio-workspace/pkg/openshift/parser"

	openshiftApi "github.com/openshift/api/template/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("template processing", func() {

	Context("parsing yaml", func() {

		It("should process role biding template using defaults", func() {
			// given
			var yaml []byte

			// when
			yaml, err := ProcessTemplate("deploy/istio-workspace/role_binding.yaml", map[string]string{"NAMESPACE": "custom-namespace"})
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(string(yaml)).To(ContainSubstring(
				`    subjects:
      - kind: ServiceAccount
        name: istio-workspace
        namespace: custom-namespace`))
		})

		It("should process operator template using defaults", func() {
			// given
			var yaml []byte

			// when
			yaml, err := ProcessTemplate("deploy/istio-workspace/operator.yaml", map[string]string{"IKE_VERSION": version.Version})
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(string(yaml)).To(Equal(processedOperatorTmplWithDefaults))
		})

		It("should process operator template using defaults and custom values", func() {
			// given
			templateValues := map[string]string{
				"IKE_DOCKER_REGISTRY":   "localhost:5000",
				"IKE_DOCKER_REPOSITORY": "ikey",
				"IKE_IMAGE_TAG":         "b1f1faf1",
			}
			var yaml []byte

			// when
			yaml, err := ProcessTemplate("deploy/istio-workspace/operator.yaml", templateValues)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(string(yaml)).To(ContainSubstring("image: localhost:5000/ikey/istio-workspace:b1f1faf1"))
		})

		Context("substituting environment variables", func() {

			It("should process operator template", func() {
				// given
				restoreEnvVars := test.TemporaryEnvVars("IKE_DOCKER_REGISTRY", "quay.io",
					"IKE_DOCKER_REPOSITORY", "istio-workspace",
					"IKE_IMAGE_NAME", "ike-cli",
					"IKE_IMAGE_TAG", "latest")
				defer restoreEnvVars()

				var yaml []byte

				// when
				yaml, err := ProcessTemplateUsingEnvVars("deploy/istio-workspace/operator.yaml")
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(string(yaml)).To(ContainSubstring("image: quay.io/istio-workspace/ike-cli:latest"))
			})
		})
	})

	Context("conversion to objects", func() {

		It("should process yaml to Openshift Template", func() {

			// when
			rawTemplate, err := ProcessTemplateUsingEnvVars("deploy/istio-workspace/operator.yaml")
			Expect(err).ToNot(HaveOccurred())

			raw, err := Parse(rawTemplate)
			Expect(err).ToNot(HaveOccurred())
			template := raw.(*openshiftApi.Template)

			// then
			Expect(template.Objects).To(HaveLen(1))
		})

	})
})

var processedOperatorTmplWithDefaults = `kind: Template
apiVersion: template.openshift.io/v1
parameters:
  - name: IKE_DOCKER_REGISTRY
    description: "Docker registry where deployed image is available"
    required: true
    value: quay.io
  - name: IKE_DOCKER_REPOSITORY
    description: "Repository in which the image can be found"
    required: true
    value: maistra
  - name: IKE_IMAGE_NAME
    description: "The name of the image with ike binary"
    required: true
    value: istio-workspace
  - name: IKE_IMAGE_TAG
    description: "The tag of the image to be used"
    required: true
    value: latest
  - name: WATCH_NAMESPACE
    description: "The namespace to watch. Leave empty for cluster wide watch."
    required: false
    value: ""
objects:
  - kind: Deployment
    apiVersion: apps/v1
    metadata:
      name: istio-workspace
      labels:
        app: istio-workspace
        version: ` + version.Version + `
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: istio-workspace
          version: ` + version.Version + `
      template:
        metadata:
          name: istio-workspace
          labels:
            app: istio-workspace
            version: ` + version.Version + `
        spec:
          serviceAccountName: istio-workspace
          containers:
            - name: istio-workspace
              image: quay.io/maistra/istio-workspace:latest
              command:
                - ike
              args:
                - serve
              imagePullPolicy: Always
              env:
                - name: WATCH_NAMESPACE
                  value: ""
                - name: POD_NAME
                  valueFrom:
                    fieldRef:
                      fieldPath: metadata.name
                - name: OPERATOR_NAME
                  value: "istio-workspace"
`
