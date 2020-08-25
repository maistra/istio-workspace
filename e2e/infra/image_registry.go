package infra

import (
	"os"

	"github.com/maistra/istio-workspace/pkg/cmd/config"

	"github.com/onsi/gomega"
)

const ImageRepo = "istio-workspace-images"

func SetDockerRegistryExternal() string {
	registry := "default-route-openshift-image-registry." + GetClusterHost()
	if externalRegistry, found := os.LookupEnv("IKE_EXTERNAL_DOCKER_REGISTRY"); found {
		registry = externalRegistry
	}
	setDockerRegistry(registry)
	return registry
}

func SetDockerRegistryInternal() {
	setDockerRegistry(GetDockerRegistryInternal())
}

func setDockerRegistry(registry string) {
	err := os.Setenv(config.EnvDockerRegistry, registry)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
}

func setDockerRepository(namespace string) {
	err := os.Setenv(config.EnvDockerRepository, namespace)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
}

// GetDockerRegistryInternal returns the internal address for the docker registry.
func GetDockerRegistryInternal() string {
	if internalRegistry, found := os.LookupEnv("IKE_INTERNAL_DOCKER_REGISTRY"); found {
		return internalRegistry
	}
	return "image-registry.openshift-image-registry.svc:5000"
}
