package main

import (
	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	istionetwork "istio.io/api/pkg/kube/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

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
					{
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
