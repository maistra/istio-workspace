package main

import (
	osappsv1 "github.com/openshift/api/apps/v1"
	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/api/pkg/kube/apis/networking/v1alpha3"
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
		}
	}
}

// Call modifier to have the test service call another. Combine with ForService
func Call(target string) Modifier {
	return func(service Entry, object runtime.Object) {
		if obj, ok := object.(*appsv1.Deployment); ok {
			obj.Spec.Template.Spec.Containers[0].Env = append(obj.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
				Name:  envServiceCall,
				Value: "http://" + target + ":9080/",
			})
		}
		if obj, ok := object.(*osappsv1.DeploymentConfig); ok {
			obj.Spec.Template.Spec.Containers[0].Env = append(obj.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
				Name:  envServiceCall,
				Value: "http://" + target + ":9080/",
			})
		}
	}
}

// ForService modifier is a filter to only execute the given modifiers if the target object belongs to the named target.
func ForService(target string, modifiers ...Modifier) Modifier {
	return func(service Entry, object runtime.Object) {
		if target != service.Name {
			return
		}
		for _, modifier := range modifiers {
			modifier(service, object)
		}
	}
}

// WithVersion modifier adds a single istio 'version' to DestinationRule/VirtualService/Deployment
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
							Host:   service.Name,
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
