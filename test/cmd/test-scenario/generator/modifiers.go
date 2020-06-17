package generator

import (
	osappsv1 "github.com/openshift/api/apps/v1"
	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ConnectToGateway modifier to connect VirtualService to a Gateway. Combine with ForService.
func ConnectToGateway() Modifier {
	return func(service Entry, object runtime.Object) {
		if obj, ok := object.(*istionetwork.VirtualService); ok {
			obj.Spec.Hosts = []string{"*"}
			obj.Spec.Gateways = append(obj.Spec.Gateways, "test-gateway")
			for i := 0; i < len(obj.Spec.Http); i++ {
				http := obj.Spec.Http[i]
				for n := 0; n < len(http.Route); n++ {
					route := http.Route[n]
					route.Destination.Port = &istiov1alpha3.PortSelector{Number: 9080}
					http.Route[n] = route
				}
				obj.Spec.Http[i] = http
			}
		}
	}
}

// GatewayOnHost modifier to set a hostname on the gateway.
func GatewayOnHost(hostname string) Modifier {
	return func(service Entry, object runtime.Object) {
		if obj, ok := object.(*istionetwork.Gateway); ok {
			for _, server := range obj.Spec.Servers {
				server.Hosts = append(server.Hosts, hostname)
			}
		}
	}
}

// Protocol is a function that returns the URL for a given Protocol for a given Service.
type Protocol func(target Entry) string

// HTTP returns the HTTP URL for the given target.
func HTTP() Protocol {
	return func(target Entry) string {
		return "http://" + target.HostName() + ":9080"
	}
}

// GRPC returns the GRPC URL for the given target.
func GRPC() Protocol {
	return func(target Entry) string {
		return target.HostName() + ":9081"
	}
}

// Call modifier to have the test service call another. Combine with ForService.
func Call(proto Protocol, target Entry) Modifier {
	return func(service Entry, object runtime.Object) {
		appendOrAdd := func(name, value string, vars []corev1.EnvVar) []corev1.EnvVar {
			found := false
			for i, envvar := range vars {
				if envvar.Name == name {
					found = true
					envvar.Value = envvar.Value + "," + value
					vars[i] = envvar
					break
				}
			}
			if !found {
				vars = append(vars, corev1.EnvVar{
					Name:  name,
					Value: value,
				})
			}
			return vars
		}

		if obj, ok := object.(*appsv1.Deployment); ok {
			obj.Spec.Template.Spec.Containers[0].Env = appendOrAdd(
				envServiceCall, proto(target),
				obj.Spec.Template.Spec.Containers[0].Env)
		}
		if obj, ok := object.(*osappsv1.DeploymentConfig); ok {
			obj.Spec.Template.Spec.Containers[0].Env = appendOrAdd(
				envServiceCall, proto(target),
				obj.Spec.Template.Spec.Containers[0].Env)
		}
	}
}

// ForService modifier is a filter to only execute the given modifiers if the target object belongs to the named target.
func ForService(target Entry, modifiers ...Modifier) Modifier {
	return func(service Entry, object runtime.Object) {
		if target.Name != service.Name {
			return
		}
		for _, modifier := range modifiers {
			modifier(service, object)
		}
	}
}

// WithVersion modifier adds a single istio 'version' to DestinationRule/VirtualService/Deployment.
func WithVersion(version string) Modifier {
	return func(service Entry, object runtime.Object) {
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
					{
						Destination: &istiov1alpha3.Destination{
							Host:   service.HostName(),
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
				if env.Name == envServiceName {
					env.Value = env.Value + "-" + version
					obj.Spec.Template.Spec.Containers[0].Env[index] = env
				}
			}
		}
		if obj, ok := object.(*osappsv1.DeploymentConfig); ok {
			obj.Spec.Template.Labels["version"] = version
			obj.ObjectMeta.Name = obj.ObjectMeta.Name + "-" + version

			for index, env := range obj.Spec.Template.Spec.Containers[0].Env {
				if env.Name == envServiceName {
					env.Value = env.Value + "-" + version
					obj.Spec.Template.Spec.Containers[0].Env[index] = env
				}
			}
		}
	}
}
