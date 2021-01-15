package testclient

import (
	"context"

	"github.com/maistra/istio-workspace/pkg/api/maistra/v1alpha1"

	"github.com/onsi/gomega"
	osappsv1 "github.com/openshift/api/apps/v1"
	istionetwork "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
 * Test Getters for Operator test suite
 */

// New returns a new set of Getters for a given Client.
func New(c client.Client) *Getters {
	return &Getters{
		Session:                   Session(c),
		Gateway:                   Gateway(c),
		DestinationRule:           DestinationRule(c),
		VirtualService:            VirtualService(c),
		VirtualServices:           VirtualServices(c),
		Deployment:                Deployment(c),
		DeploymentWithError:       DeploymentWithError(c),
		DeploymentConfig:          DeploymentConfig(c),
		DeploymentConfigWithError: DeploymentConfigWithError(c),
	}
}

// Getters simple struct to hold funcs.
type Getters struct {
	Session                   func(namespace, name string) v1alpha1.Session
	Gateway                   func(namespace, name string) istionetwork.Gateway
	DestinationRule           func(namespace, name string) istionetwork.DestinationRule
	VirtualService            func(namespace, name string) istionetwork.VirtualService
	Deployment                func(namespace, name string) appsv1.Deployment
	DeploymentWithError       func(namespace, name string) (appsv1.Deployment, error)
	DeploymentConfig          func(namespace, name string) osappsv1.DeploymentConfig
	DeploymentConfigWithError func(namespace, name string) (osappsv1.DeploymentConfig, error)
	VirtualServices           func(namespace string) istionetwork.VirtualServiceList
}

// Session returns a session by name in a given namespace.
func Session(c client.Client) func(namespace, name string) v1alpha1.Session {
	return func(namespace, name string) v1alpha1.Session {
		s := v1alpha1.Session{}
		err := c.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &s)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		return s
	}
}

// Gateway returns a gateway by name in a given namespace.
func Gateway(c client.Client) func(namespace, name string) istionetwork.Gateway {
	return func(namespace, name string) istionetwork.Gateway {
		s := istionetwork.Gateway{}
		err := c.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &s)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		return s
	}
}

// DestinationRule returns a destinationrule by name in a given namespace.
func DestinationRule(c client.Client) func(namespace, name string) istionetwork.DestinationRule {
	return func(namespace, name string) istionetwork.DestinationRule {
		s := istionetwork.DestinationRule{}
		err := c.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &s)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		return s
	}
}

// VirtualService returns a virtualservice by name in a given namespace.
func VirtualService(c client.Client) func(namespace, name string) istionetwork.VirtualService {
	return func(namespace, name string) istionetwork.VirtualService {
		s := istionetwork.VirtualService{}
		err := c.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &s)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		return s
	}
}

// Deployment returns a deployment by name in a given namespace.
func Deployment(c client.Client) func(namespace, name string) appsv1.Deployment {
	return func(namespace, name string) appsv1.Deployment {
		s := appsv1.Deployment{}
		err := c.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &s)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		return s
	}
}

// DeploymentWithError returns a deployment by name in a given namespace or error.
func DeploymentWithError(c client.Client) func(namespace, name string) (appsv1.Deployment, error) {
	return func(namespace, name string) (appsv1.Deployment, error) {
		s := appsv1.Deployment{}
		err := c.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &s)
		return s, err
	}
}

// DeploymentConfig returns a deploymentconfig by name in a given namespace.
func DeploymentConfig(c client.Client) func(namespace, name string) osappsv1.DeploymentConfig {
	return func(namespace, name string) osappsv1.DeploymentConfig {
		s := osappsv1.DeploymentConfig{}
		err := c.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &s)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		return s
	}
}

// DeploymentConfigWithError returns a deploymentconfig by name in a given namespace or error.
func DeploymentConfigWithError(c client.Client) func(namespace, name string) (osappsv1.DeploymentConfig, error) {
	return func(namespace, name string) (osappsv1.DeploymentConfig, error) {
		s := osappsv1.DeploymentConfig{}
		err := c.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &s)
		return s, err
	}
}

// VirtualServices returns all virtualservices in a given namespace.
func VirtualServices(c client.Client) func(namespace string) istionetwork.VirtualServiceList {
	return func(namespace string) istionetwork.VirtualServiceList {
		s := istionetwork.VirtualServiceList{}
		err := c.List(context.Background(), &s, client.InNamespace(namespace))
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		return s
	}
}
