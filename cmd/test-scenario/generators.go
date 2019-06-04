package main

import (
	"fmt"

	"sigs.k8s.io/yaml"

	osappsv1 "github.com/openshift/api/apps/v1"
	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/api/pkg/kube/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Modifier func(service string, object runtime.Object)
type SubGenerator func(service string) runtime.Object

func Generate(services []string, modifiers ...Modifier) {

	sub := []SubGenerator{Deployment, Service, DestinationRule, VirtualService}
	modify := func(service string, object runtime.Object) {
		for _, modifier := range modifiers {
			modifier(service, object)
		}
	}
	printObj := func(object runtime.Object) {
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
				printObj(object)
			}
		}(service)
	}
	gw := Gateway()
	modify("gateway", gw)
	printObj(gw)
}

func DeploymentConfig(service string) runtime.Object {
	return &osappsv1.DeploymentConfig{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "DeploymentConfig",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: service,
		},
		Spec: osappsv1.DeploymentConfigSpec{
			Replicas: 1,
			Template: &corev1.PodTemplateSpec{
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
						{
							Name:            service,
							Image:           "aslakknutsen/istio-workspace-test:latest",
							ImagePullPolicy: "Always",
							Env: []corev1.EnvVar{
								{
									Name:  "SERVICE_NAME",
									Value: service,
								},
								{
									Name:  "HTTP_ADDR",
									Value: ":9080",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 9080,
								},
							},
							/*
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
							*/
						},
					},
				},
			},
		},
	}
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
						{
							Name:            service,
							Image:           "aslakknutsen/istio-workspace-test:latest",
							ImagePullPolicy: "Always",
							Env: []corev1.EnvVar{
								{
									Name:  "SERVICE_NAME",
									Value: service,
								},
								{
									Name:  "HTTP_ADDR",
									Value: ":9080",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 9080,
								},
							},
							/*
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
							*/
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
				{
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
				{
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
