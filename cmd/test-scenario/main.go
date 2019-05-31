package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"

	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/api/pkg/kube/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("required arg 'scenario name' missing")
		os.Exit(-100)
	}
	scenarios := map[string]func(){
		"scenario-1": TestScenario1,
	}
	scenario := os.Args[1]
	if f, ok := scenarios[scenario]; ok {
		f()

	} else {
		fmt.Println("Scenario not found", scenario)
		os.Exit(-101)
	}
}

func TestScenario1() {
	// Scenario 1
	services := []string{"productpage", "reviews", "ratings"}
	Generate(
		services,
		WithVersion("v1"),
		ForService("productpage", Call("reviews"), ConnectToGatway()),
		ForService("reviews", Call("ratings")),
	)
}

func ConnectToGatway() Modifier {
	return func(service string, object runtime.Object) {
		if obj, ok := object.(*istionetwork.VirtualService); ok {
			obj.Spec.Hosts = []string{"*"}
			obj.Spec.Gateways = append(obj.Spec.Gateways, "test-gateway")
		}
	}
}

func Call(target string) Modifier {
	return func(service string, object runtime.Object) {
		if obj, ok := object.(*appsv1.Deployment); ok {
			obj.Spec.Template.Spec.Containers[0].Env = append(obj.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
				Name:  "SERVICE_CALL",
				Value: "http://" + target + ":9080/",
			})
		}
	}
}
func ForService(target string, modifiers ...Modifier) Modifier {
	return func(service string, object runtime.Object) {
		if target != service {
			return
		}
		for _, modifier := range modifiers {
			modifier(service, object)
		}
	}
}

func WithVersion(version string) Modifier {
	return func(service string, object runtime.Object) {
		if obj, ok := object.(*istionetwork.DestinationRule); ok {
			obj.Spec.Subsets = append(obj.Spec.Subsets, &istiov1alpha3.Subset{
				Name: version,
				Labels: map[string]string{
					"version": version,
				},
			})
		}
		if obj, ok := object.(*istionetwork.VirtualService); ok {
			obj.Spec.Http = append(obj.Spec.Http, &istiov1alpha3.HTTPRoute{
				Route: []*istiov1alpha3.HTTPRouteDestination{
					&istiov1alpha3.HTTPRouteDestination{
						Destination: &istiov1alpha3.Destination{
							Host:   service,
							Subset: version,
						},
					},
				},
			})
		}
		if obj, ok := object.(*appsv1.Deployment); ok {
			obj.Spec.Template.Labels["version"] = version
			obj.ObjectMeta.Name = obj.ObjectMeta.Name + "-" + version

			for index, env := range obj.Spec.Template.Spec.Containers[0].Env {
				if env.Name == "SERVICE_NAME" {
					env.Value = env.Value + "-" + version
					obj.Spec.Template.Spec.Containers[0].Env[index] = env
				}
			}
		}
	}
}

type Modifier func(service string, object runtime.Object)
type SubGenerator func(service string) runtime.Object

func Generate(services []string, modifiers ...Modifier) {

	sub := []SubGenerator{Deployment, Service, DestinationRule, VirtualService}
	modify := func(service string, object runtime.Object) {
		for _, modifier := range modifiers {
			modifier(service, object)
		}
	}
	print := func(object runtime.Object) {
		b, err := yaml.Marshal(object)
		if err != nil {
			fmt.Println("Marshal error", err)
		}
		fmt.Println(string(b))
		fmt.Println("---")
	}
	for _, service := range services {
		func(service string) {
			for _, subGenerator := range sub {
				object := subGenerator(service)
				if object == nil {
					continue
				}
				modify(service, object)
				print(object)
			}
		}(service)
	}
	gw := Gateway()
	modify("gateway", gw)
	print(gw)
}

func Deployment(service string) runtime.Object {
	replica := int32(1)
	return &appsv1.Deployment{
		TypeMeta: v1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Deployment",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: service,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replica,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "true",
						"prometheus.io/scrape":    "true",
						"prometheus.io/port":      "9080",
						"prometheus.io/scheme":    "http",
						"prometheus.io/path":      "/metrics",
						"kiali.io/runtimes":       "go",
					},
					Labels: map[string]string{
						"app": service,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:            service,
							Image:           "aslakknutsen/istio-workspace-test:latest",
							ImagePullPolicy: "Always",
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:  "SERVICE_NAME",
									Value: service,
								},
								corev1.EnvVar{
									Name:  "HTTP_ADDR",
									Value: ":9080",
								},
							},
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									ContainerPort: 9080,
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: intstr.FromInt(9080),
									},
								},
								InitialDelaySeconds: 1,
								PeriodSeconds:       3,
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: intstr.FromInt(9080),
									},
								},
								InitialDelaySeconds: 1,
								PeriodSeconds:       3,
							},
						},
					},
				},
			},
		},
	}
}

func Service(service string) runtime.Object {
	return &corev1.Service{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: service,
			Labels: map[string]string{
				"app": service,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name: "http",
					Port: 9080,
				},
			},
			Selector: map[string]string{
				"app": service,
			},
		},
	}
}

func DestinationRule(service string) runtime.Object {
	return &istionetwork.DestinationRule{
		TypeMeta: v1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "DestinationRule",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: service,
		},
		Spec: istiov1alpha3.DestinationRule{
			Host: service,
		},
	}
}

func VirtualService(service string) runtime.Object {
	return &istionetwork.VirtualService{
		TypeMeta: v1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "VirtualService",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: service,
		},
		Spec: istiov1alpha3.VirtualService{
			Hosts: []string{service},
		},
	}
}

func Gateway() runtime.Object {
	return &istionetwork.Gateway{
		TypeMeta: v1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "Gateway",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "test-gateway",
		},
		Spec: istiov1alpha3.Gateway{
			Selector: map[string]string{
				"istio": "ingressgateway",
			},
			Servers: []*istiov1alpha3.Server{
				&istiov1alpha3.Server{
					Port: &istiov1alpha3.Port{
						Protocol: "HTTP",
						Name:     "http",
						Number:   80,
					},
					Hosts: []string{"*"},
				},
			},
		},
	}
}
