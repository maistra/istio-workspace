package dynclient_test

import (
	"os"

	"github.com/maistra/istio-workspace/pkg/openshift/dynclient"

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
			yaml, err := dynclient.ProcessTemplate("deploy/istio-workspace/role_binding.yaml", map[string]string{"NAMESPACE": "custom-namespace"})
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
			yaml, err := dynclient.ProcessTemplate("deploy/istio-workspace/operator.yaml", nil)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(string(yaml)).To(Equal(processedTmplWithDefaults))
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
			yaml, err := dynclient.ProcessTemplate("deploy/istio-workspace/operator.yaml", templateValues)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(string(yaml)).To(ContainSubstring("image: localhost:5000/ikey/istio-workspace:b1f1faf1"))
		})

		Context("substituting environment variables", func() {

			It("should process operator template", func() {
				// given
				newEnvs := map[string]string{
					"IKE_DOCKER_REGISTRY":   "quay.io",
					"IKE_DOCKER_REPOSITORY": "istio-workspace",
					"IKE_IMAGE_NAME":        "ike-cli",
					"IKE_IMAGE_TAG":         "latest",
				}

				oldEnvs := map[string]string{}
				for k, v := range newEnvs {
					oldEnvs[os.Getenv(k)] = v
					_ = os.Setenv(k, v)
				}
				defer func() {
					for k, v := range oldEnvs {
						_ = os.Setenv(k, v)
					}
				}()

				var yaml []byte

				// when
				yaml, err := dynclient.ProcessTemplateUsingEnvVars("deploy/istio-workspace/operator.yaml")
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(string(yaml)).To(ContainSubstring("image: quay.io/istio-workspace/ike-cli:latest"))
			})
		})
	})

	Context("conversion to objects", func() {

		It("should process yaml to Openshift Template", func() {

			// when
			rawTemplate, err := dynclient.ProcessTemplateUsingEnvVars("deploy/istio-workspace/operator.yaml")
			Expect(err).ToNot(HaveOccurred())

			raw, err := dynclient.Parse(rawTemplate)
			Expect(err).ToNot(HaveOccurred())
			template := raw.(*openshiftApi.Template)

			// then
			Expect(template.Objects).To(HaveLen(1))
		})

	})
})

const processedTmplWithDefaults = `kind: Template
apiVersion: template.openshift.io/v1
parameters:
  - name: IKE_DOCKER_REGISTRY
    required: true
    value: docker.io
  - name: IKE_DOCKER_REPOSITORY
    required: true
    value: aslakknutsen
  - name: IKE_IMAGE_NAME
    required: true
    value: istio-workspace
  - name: IKE_IMAGE_TAG
    required: true
    value: latest
  - name: TELEPRESENCE_VERSION
    required: true
    value: "0.100"
objects:
  - kind: Deployment
    apiVersion: apps/v1
    metadata:
      name: istio-workspace
    spec:
      replicas: 1
      selector:
        matchLabels:
          name: istio-workspace
      template:
        metadata:
          labels:
            name: istio-workspace
        spec:
          serviceAccountName: istio-workspace
          containers:
            - name: istio-workspace
              image: docker.io/aslakknutsen/istio-workspace:latest
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
                - name: TELEPRESENCE_VERSION
                  value: "0.100"
`
