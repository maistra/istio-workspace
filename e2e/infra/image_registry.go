package infra

import (
	"os"

	"github.com/onsi/ginkgo"

	"github.com/maistra/istio-workspace/pkg/cmd/config"

	"github.com/onsi/gomega"
)

// GetRepositoryName returns the name of the repository http://host/repository-name/image-name:tag
func GetRepositoryName() string {
	if UsePrebuiltImages() {
		if dockerRegistry, found := os.LookupEnv("IKE_DOCKER_REPOSITORY"); !found {
			ginkgo.Fail("\"IKE_DOCKER_REPOSITORY\" env variable not set")
		} else {
			return dockerRegistry
		}
	}
	// used to reuse images pushed to a single namespace to avoid rebuilding pr test
	return "istio-workspace-images"
}

// GetImageTag returns image tag if defined in IKE_IMAGE_TAG variable or "latest" otherwise.
func GetImageTag() string {
	if imageTag, found := os.LookupEnv("IKE_IMAGE_TAG"); found {
		return imageTag
	}

	return "latest"
}

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
