package template

import (
	"fmt"
	"testing"
)

func TestTemplateJSONValue(t *testing.T) {
	tj, err := newJSON([]byte(org))
	if err != nil {
		t.Error(err)
	}
	if !tj.Has("/metadata/creationTimestamp") {
		t.Error("creationTimestamp not found")
	}
	if v, _ := tj.Value("/spec/replicas"); v.(float64) != 1 {
		t.Error("replicas not found")
	}

	if v, _ := tj.Value("/metadata/labels/version"); v.(string) != "v1" {
		t.Error("version not found")
	}

	if !tj.Equal("/metadata/labels/version", "v1") {
		t.Error("version not equal")
	}

	if v, _ := tj.Value("/spec/template/spec/containers/0/env/0/value"); v.(string) != "productpage-v1" {
		t.Error("env not found")
	}
}

func TestX(t *testing.T) {

	e := NewDefaultEngine()

	modified, err := e.Run("telepresence", []byte(org), "1000", map[string]string{
		"TelepresenceVersion": "x",
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Modified document: %s\n", modified)
}

var org = `
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
