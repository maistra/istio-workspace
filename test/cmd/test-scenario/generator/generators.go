package generator

import (
	"io"
	"time"

	"sigs.k8s.io/yaml"

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
	TestImageName = ""
	GatewayHost   = "*"
)

// Entry is a simple value object that holds the basic configuration used by the generator.
type Entry struct {
	Name           string
	DeploymentType string
	Namespace      string
}

// HostName return the full cluster host name if Namespace is set or the local if not.
func (e *Entry) HostName() string {
	if e.Namespace != "" {
		return e.Name + "." + e.Namespace + ".svc.cluster.local"
	}
	return e.Name
}

// SubGenerator is a function intended to create the basic runtime.Object as a starting point for modification.
type SubGenerator func(service Entry) runtime.Object

// Modifier is a function to change a runtime.Object into something more specific for a given scenario.
type Modifier func(service Entry, object runtime.Object)

// Generate runs and prints the full test scenario generation to sysout.
func Generate(out io.Writer, services []Entry, modifiers ...Modifier) {
	sub := []SubGenerator{Deployment, DeploymentConfig, Service, DestinationRule, VirtualService}
	modify := func(service Entry, object runtime.Object) {
		for _, modifier := range modifiers {
			modifier(service, object)
		}
	}
	printObj := func(object runtime.Object) {
		b, err := yaml.Marshal(object)
		if err != nil {
			_, _ = io.WriteString(out, "Marshal error"+err.Error()+"\n")
		}
		_, _ = out.Write(b)
		_, _ = io.WriteString(out, "---\n")
	}
	for _, service := range services {
		func(service Entry) {
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
	modify(Entry{Name: "gateway"}, gw)
	printObj(gw)
}

// DeploymentConfig basic SubGenerator for the kind DeploymentConfig.
func DeploymentConfig(service Entry) runtime.Object {
	if service.DeploymentType != "DeploymentConfig" {
		return nil
	}
	template := template(service.Name)
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
func Deployment(service Entry) runtime.Object {
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
			Template: template(service.Name),
		},
	}
}

// Service basic SubGenerator for the kind Service.
func Service(service Entry) runtime.Object {
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
					Port: 9080,
				},
				{
					Name: "grpc",
					Port: 9081,
				},
			},
			Selector: map[string]string{
				"app": service.Name,
			},
		},
	}
}

// DestinationRule basic SubGenerator for the kind DestinationRule.
func DestinationRule(service Entry) runtime.Object {
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
func VirtualService(service Entry) runtime.Object {
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
			Http: []*istiov1alpha3.HTTPRoute{
				{
					Match: []*istiov1alpha3.HTTPMatchRequest{
						{
							Uri: &istiov1alpha3.StringMatch{
								MatchType: &istiov1alpha3.StringMatch_Prefix{
									Prefix: "/test-service",
								},
							},
						},
					},
					Rewrite: &istiov1alpha3.HTTPRewrite{
						Uri: "/",
					},
					Route: []*istiov1alpha3.HTTPRouteDestination{
						{
							Destination: &istiov1alpha3.Destination{
								Host: service.HostName(),
							},
						},
					},
				},
			},
		},
	}
}

// Gateway basic SubGenerator for the kind Gateway.
func Gateway() runtime.Object {
	return &istionetwork.Gateway{
		TypeMeta: v1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "Gateway",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-gateway",
			Namespace: Namespace,
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

func template(name string) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
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
				"app": name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            name,
					Image:           TestImageName,
					ImagePullPolicy: "Always",
					Env: []corev1.EnvVar{
						{
							Name:  envServiceName,
							Value: name,
						},
						{
							Name:  "HTTP_ADDR",
							Value: ":9080",
						},
						{
							Name:  "GRPC_ADDR",
							Value: ":9081",
						},
					},
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: 9080,
						},
						{
							Name:          "grpc",
							ContainerPort: 9081,
						},
					},
					LivenessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/healthz",
								Port: intstr.FromInt(9080),
							},
						},
						InitialDelaySeconds: 5,
						PeriodSeconds:       3,
					},
					ReadinessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/healthz",
								Port: intstr.FromInt(9080),
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
