package generator

import (
	"fmt"
	"time"

	osappsv1 "github.com/openshift/api/apps/v1"
	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	envServiceName = "SERVICE_NAME"
	envServiceCall = "SERVICE_CALL"
)

var (
	GatewayHost      = "*"
	NsGenerators     = []SubGenerator{Gateway}
	AllSubGenerators = []SubGenerator{Deployment, DeploymentConfig, Service, DestinationRule, VirtualService}
)

// ServiceEntry is a simple value object that holds the basic configuration used by the generator.
type ServiceEntry struct {
	Name           string
	DeploymentType string
	Image          string
	Namespace      string
	Gateway        string
	HTTPPort       uint32
	GRPCPort       uint32
}

func NewServiceEntry(name, namespace, deploymentType, image string) ServiceEntry {
	return ServiceEntry{Name: name,
		Namespace:      namespace,
		DeploymentType: deploymentType,
		Image:          image,
		Gateway:        "test-gateway",
		HTTPPort:       9080,
		GRPCPort:       9081}
}

// HostName return the full cluster host name if Namespace is set or the local if not.
func (e *ServiceEntry) HostName() string {
	if e.Namespace != "" {
		return e.Name + "." + e.Namespace + ".svc.cluster.local"
	}

	return e.Name
}

// SubGenerator is a function intended to create the basic runtime.Object as a starting point for modification.
type SubGenerator func(service ServiceEntry) runtime.Object

// Modifier is a function to change a runtime.Object into something more specific for a given scenario.
type Modifier func(service ServiceEntry, object runtime.Object)

// Generate runs and prints the full test scenario generation to sysout.
func Generate(printer Printer, services []ServiceEntry, gen, sub []SubGenerator, modifiers ...Modifier) {
	modify := func(service ServiceEntry, object runtime.Object) {
		for _, modifier := range modifiers {
			modifier(service, object)
		}
	}

	// These generators run once per namespace as they construct unique resources.
	// Assumption: service entries holds ns-specific data unified
	// e.g. gateway is always the same
	for _, generator := range gen {
		gw := generator(services[0])
		modify(ServiceEntry{Gateway: services[0].Gateway}, gw)
		printer(gw)
	}

	for _, service := range services {
		func(service ServiceEntry) {
			for _, subGenerator := range sub {
				object := subGenerator(service)
				if object == nil {
					continue
				}
				modify(service, object)
				printer(object)
			}
		}(service)
	}
}

// DeploymentConfig basic SubGenerator for the kind DeploymentConfig.
func DeploymentConfig(service ServiceEntry) runtime.Object {
	if service.DeploymentType != "DeploymentConfig" {
		return nil
	}
	template := template(service)

	return &osappsv1.DeploymentConfig{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "DeploymentConfig",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      service.Name,
			Namespace: service.Namespace,
			Labels: map[string]string{
				"app": service.Name,
			},
			CreationTimestamp: v1.Time{Time: time.Now()},
		},
		Spec: osappsv1.DeploymentConfigSpec{
			Replicas: 1,
			Template: &template,
		},
	}
}

// Deployment basic SubGenerator for the kind Deployment.
func Deployment(service ServiceEntry) runtime.Object {
	if service.DeploymentType != "Deployment" {
		return nil
	}
	replica := int32(1)

	return &appsv1.Deployment{
		TypeMeta: v1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:              service.Name,
			Namespace:         service.Namespace,
			CreationTimestamp: v1.Time{Time: time.Now()},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replica,
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					"app": service.Name,
				},
			},
			Template: template(service),
		},
	}
}

// Service basic SubGenerator for the kind Service.
func Service(service ServiceEntry) runtime.Object {
	return &corev1.Service{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      service.Name,
			Namespace: service.Namespace,
			Labels: map[string]string{
				"app": service.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: int32(service.HTTPPort),
				},
				{
					Name: "grpc",
					Port: int32(service.GRPCPort),
				},
			},
			Selector: map[string]string{
				"app": service.Name,
			},
		},
	}
}

// DestinationRule basic SubGenerator for the kind DestinationRule.
func DestinationRule(service ServiceEntry) runtime.Object {
	return &istionetwork.DestinationRule{
		TypeMeta: v1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "DestinationRule",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      service.Name,
			Namespace: service.Namespace,
		},
		Spec: istiov1alpha3.DestinationRule{
			Host: service.HostName(),
		},
	}
}

// VirtualService basic SubGenerator for the kind VirtualService.
func VirtualService(service ServiceEntry) runtime.Object {
	return &istionetwork.VirtualService{
		TypeMeta: v1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "VirtualService",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      service.Name,
			Namespace: service.Namespace,
		},
		Spec: istiov1alpha3.VirtualService{
			Hosts: []string{service.HostName()},
			Http:  []*istiov1alpha3.HTTPRoute{},
		},
	}
}

// Gateway basic SubGenerator for the kind Gateway.
func Gateway(service ServiceEntry) runtime.Object {
	return &istionetwork.Gateway{
		TypeMeta: v1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "Gateway",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      service.Gateway,
			Namespace: service.Namespace,
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
					Hosts: []string{},
				},
			},
		},
	}
}

func template(service ServiceEntry) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				"sidecar.istio.io/inject": "true",
				"prometheus.io/scrape":    "true",
				"prometheus.io/port":      fmt.Sprintf("%d", service.HTTPPort),
				"prometheus.io/scheme":    "http",
				"prometheus.io/path":      "/metrics",
				"kiali.io/runtimes":       "go",
			},
			Labels: map[string]string{
				"app": service.Name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            service.Name,
					Image:           service.Image, // FIX take from Service entry?
					ImagePullPolicy: "Always",
					Env: []corev1.EnvVar{
						{
							Name:  envServiceName,
							Value: service.Name,
						},
						{
							Name:  "HTTP_ADDR",
							Value: fmt.Sprintf(":%d", service.HTTPPort),
						},
						{
							Name:  "GRPC_ADDR",
							Value: fmt.Sprintf(":%d", service.GRPCPort),
						},
					},
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: int32(service.HTTPPort),
						},
						{
							Name:          "grpc",
							ContainerPort: int32(service.GRPCPort),
						},
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/healthz",
								Port: intstr.FromInt(int(service.HTTPPort)),
							},
						},
						InitialDelaySeconds: 5,
						PeriodSeconds:       3,
						FailureThreshold:    10,
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/healthz",
								Port: intstr.FromInt(int(service.HTTPPort)),
							},
						},
						InitialDelaySeconds: 5,
						PeriodSeconds:       3,
					},
				},
			},
		},
	}
}
