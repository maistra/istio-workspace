package template_test

import (
	"github.com/maistra/istio-workspace/pkg/template"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Operations for template system", func() {

	Context("json object", func() {
		var (
			err error
			tj  template.JSON
		)

		BeforeEach(func() {
			tj, err = template.NewJSON([]byte(testDeployment))
			Expect(err).ToNot(HaveOccurred())
		})

		Context("map values", func() {
			It("should error on missing root path", func() {
				_, err := tj.Value("")
				Expect(err).To(HaveOccurred())
			})
			It("should have nil value in missing parent", func() {
				v, err := tj.Value("/metadata/UNKNOWN/UNKNOWN2")
				Expect(err).ToNot(HaveOccurred())
				Expect(v).To(BeNil())
			})
			It("should be able to check value in map", func() {
				v := tj.Has("/metadata/creationTimestamp")
				Expect(v).To(BeTrue())
			})
			It("should be able to get numeric value", func() {
				v, err := tj.Value("/spec/replicas")
				Expect(err).ToNot(HaveOccurred())
				Expect(v).To(BeEquivalentTo(1))
			})
			It("should be able to get string value", func() {
				v, err := tj.Value("/metadata/labels/version")
				Expect(err).ToNot(HaveOccurred())
				Expect(v).To(BeEquivalentTo("v1"))
			})
			It("should be able to equal numberic values", func() {
				v := tj.Equal("/spec/replicas", 1)
				Expect(v).To(BeTrue())
			})
			It("should be able to equal string values", func() {
				v := tj.Equal("/metadata/labels/version", "v1")
				Expect(v).To(BeTrue())
			})

			It("should not equal found value", func() {
				v := tj.Equal("/metadata/labels/version", "UNKNOWN")
				Expect(v).To(BeFalse())
			})
			It("should not equal missing value", func() {
				v := tj.Equal("/metadata/labels/version-UNKNOWN", "v1")
				Expect(v).To(BeFalse())
			})
			It("should not have missing value", func() {
				v := tj.Has("/metadata/creationTimestamp-UNKNOWN")
				Expect(v).To(BeFalse())
			})
		})
		Context("slice values", func() {
			It("should have value in slice", func() {
				v := tj.Has("/spec/template/spec/containers/0")
				Expect(v).To(BeTrue())
			})
			It("should get value in slice", func() {
				v, err := tj.Value("/spec/template/spec/containers/0")
				Expect(err).ToNot(HaveOccurred())
				Expect(v).ToNot(BeEmpty())
			})
			It("should get value in child of slice", func() {
				v, err := tj.Value("/spec/template/spec/containers/0/env/0/value")
				Expect(err).ToNot(HaveOccurred())
				Expect(v).To(BeEquivalentTo("productpage-v1"))
			})
			It("should equal value in child of slice", func() {
				v := tj.Equal("/spec/template/spec/containers/0/env/0/value", "productpage-v1")
				Expect(v).To(BeTrue())
			})
			It("should have value in child of slice", func() {
				v := tj.Has("/spec/template/spec/containers/0/env/0/value")
				Expect(v).To(BeTrue())
			})

			It("should error on non numberic slice index", func() {
				_, err := tj.Value("/spec/template/spec/containers/X")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("engine", func() {

		Context("telepresence", func() {
			It("happy, happy, basic DefaultEngine", func() {
				e := template.NewDefaultEngine()

				o, err := e.Run("telepresence", []byte(testDeployment), "1000", map[string]string{
					"version": "x-x-v",
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(string(o)).To(ContainSubstring("1000"))
				Expect(string(o)).To(ContainSubstring("x-x-v"))
			})

			It("should fail when no version is provided", func() {
				e := template.NewDefaultEngine()

				_, err := e.Run("telepresence", []byte(testDeployment), "1000", map[string]string{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("expected version variable to be set"))
			})
		})

		Context("preparedimage", func() {
			It("happy, happy, basic DefaultEngine", func() {
				e := template.NewDefaultEngine()

				o, err := e.Run("prepared-image", []byte(testDeployment), "1000", map[string]string{
					"image": "maistra.org/test-image:test",
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(string(o)).To(ContainSubstring("1000"))
				Expect(string(o)).To(ContainSubstring("maistra.org/test-image:test"))
			})
		})

		Context("object validation", func() {
			It("should fail on wrong Patch format", func() {
				e := template.NewEngine(template.Patches{template.Patch{
					Name:     "test",
					Template: []byte("{"),
				}})
				_, err := e.Run("test", []byte("{}"), "x", map[string]string{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unexpected end of JSON input"))
			})

			It("should fail on wrong Patch format", func() {
				e := template.NewEngine(template.Patches{template.Patch{
					Name:     "test",
					Template: []byte("[]"),
				}})
				_, err := e.Run("test", []byte("{"), "x", map[string]string{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unexpected end of JSON input"))
			})
		})

		Context("normal operations", func() {

			It("should apply patch", func() {
				e := template.NewEngine(template.Patches{template.Patch{
					Name:      "test",
					Template:  []byte(`[ {"op": "remove", "path": "/version"} ]`),
					Variables: map[string]string{},
				}})
				o, err := e.Run("test", []byte(`{"version": "100"}`), "x", map[string]string{})
				Expect(err).ToNot(HaveOccurred())
				Expect(string(o)).ToNot(ContainSubstring("version"))
			})

			It("should fail on bad patch", func() {
				e := template.NewEngine(template.Patches{template.Patch{
					Name:      "test",
					Template:  []byte(`[ {"op": "remove", "path": "/test"} ]`),
					Variables: map[string]string{},
				}})
				_, err := e.Run("test", []byte(`{"version": "100"}`), "x", map[string]string{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to remove nonexistent key: test"))
			})
		})

		Context("variables", func() {

			It("should use default values if non provided", func() {
				e := template.NewEngine(template.Patches{template.Patch{
					Name:     "test",
					Template: []byte(`[ {"op": "replace", "path": "/version", "value": "{{.Vars.Version}}"} ]`),
					Variables: map[string]string{
						"Version": "DEFAULT_VERSION",
					},
				}})
				o, err := e.Run("test", []byte(`{"version": "100"}`), "x", map[string]string{})
				Expect(err).ToNot(HaveOccurred())
				Expect(string(o)).To(ContainSubstring("DEFAULT_VERSION"))
			})

			It("should override with incoming if available", func() {
				e := template.NewEngine(template.Patches{template.Patch{
					Name:     "test",
					Template: []byte(`[ {"op": "replace", "path": "/version", "value": "{{.Vars.Version}}"} ]`),
					Variables: map[string]string{
						"Version": "DEFAULT_VERSION",
					},
				}})
				o, err := e.Run("test", []byte(`{"version": "100"}`), "x", map[string]string{
					"Version": "PROVIDED_VERSION",
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(string(o)).To(ContainSubstring("PROVIDED_VERSION"))
			})
		})
	})
})

var testDeployment = `
{
    "apiVersion": "extensions/v1beta1",
    "kind": "Deployment",
    "metadata": {
        "annotations": {
            "deployment.kubernetes.io/revision": "1",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"extensions/v1beta1\"}\n"
        },
        "creationTimestamp": "2019-07-13T08:46:46Z",
        "generation": 1,
        "labels": {
            "app": "productpage",
            "version": "v1"
        },
        "name": "productpage-v1",
        "namespace": "bookinfo",
        "resourceVersion": "638482",
        "selfLink": "/apis/extensions/v1beta1/namespaces/bookinfo/deployments/productpage-v1",
        "uid": "bf2a3655-a54a-11e9-b309-482ae3045b54"
    },
    "spec": {
        "progressDeadlineSeconds": 600,
        "replicas": 1,
        "revisionHistoryLimit": 10,
        "selector": {
                "app": "productpage",
                "version": "v1"
        },
        "strategy": {
            "rollingUpdate": {
                "maxSurge": 1,
                "maxUnavailable": 1
            },
            "type": "RollingUpdate"
        },
        "template": {
            "metadata": {
                "annotations": {
                    "kiali.io/runtimes": "go",
                    "prometheus.io/path": "/metrics",
                    "prometheus.io/port": "9080",
                    "prometheus.io/scheme": "http",
                    "prometheus.io/scrape": "true",
                    "sidecar.istio.io/inject": "true"
                },
                "creationTimestamp": null,
                "labels": {
                    "app": "productpage",
                    "version": "v1"
                }
            },
            "spec": {
                "containers": [
                    {
                        "env": [
                            {
                                "name": "SERVICE_NAME",
                                "value": "productpage-v1"
                            },
                            {
                                "name": "HTTP_ADDR",
                                "value": ":9080"
                            },
                            {
                                "name": "SERVICE_CALL",
                                "value": "http://reviews:9080/"
                            }
                        ],
                        "image": "docker.io/aslakknutsen/istio-workspace-test:latest",
                        "imagePullPolicy": "Always",
                        "livenessProbe": {
                            "failureThreshold": 3,
                            "httpGet": {
                                "path": "/healthz",
                                "port": 9080,
                                "scheme": "HTTP"
                            },
                            "initialDelaySeconds": 1,
                            "periodSeconds": 3,
                            "successThreshold": 1,
                            "timeoutSeconds": 1
                        },
                        "name": "productpage",
                        "ports": [
                            {
                                "containerPort": 9080,
                                "protocol": "TCP"
                            }
                        ],
                        "readinessProbe": {
                            "failureThreshold": 3,
                            "httpGet": {
                                "path": "/healthz",
                                "port": 9080,
                                "scheme": "HTTP"
                            },
                            "initialDelaySeconds": 1,
                            "periodSeconds": 3,
                            "successThreshold": 1,
                            "timeoutSeconds": 1
                        },
                        "resources": {},
                        "terminationMessagePath": "/dev/termination-log",
                        "terminationMessagePolicy": "File"
                    }
                ],
                "dnsPolicy": "ClusterFirst",
                "restartPolicy": "Always",
                "schedulerName": "default-scheduler",
                "securityContext": {},
                "terminationGracePeriodSeconds": 30
            }
        }
    }
}
`
