package infra

import (
	"os"

	"github.com/maistra/istio-workspace/pkg/cmd/config"

	"github.com/onsi/gomega"
)

const ImageRepo = "istio-workspace-images"

func setDockerRegistryExternal() string {
	registry := "default-route-openshift-image-registry." + GetClusterHost()
	setDockerRegistry(registry)
	return registry
}

func setDockerRegistryInternal() {
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
	return "image-registry.openshift-image-registry.svc:5000"
}
