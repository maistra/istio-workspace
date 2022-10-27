package infra

import (
	"os"

	"github.com/maistra/istio-workspace/pkg/cmd/config"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// GetRepositoryName returns the name of the repository.
func GetRepositoryName() string {
	if UsePrebuiltImages() {
		if repository, found := os.LookupEnv("IKE_CONTAINER_REPOSITORY"); !found {
			ginkgo.Fail("\"IKE_CONTAINER_REPOSITORY\" env variable not set")
		} else {
			return repository
		}
	}
	// used to reuse images pushed to a single namespace to avoid rebuilding pr test
	return "istio-workspace-images"
}

// GetDevRepositoryName returns the name of the repository containing development related images.
func GetDevRepositoryName() string {
	if UsePrebuiltImages() {
		if repository, found := os.LookupEnv("IKE_CONTAINER_DEV_REPOSITORY"); !found {
			ginkgo.Fail("\"IKE_CONTAINER_DEV_REPOSITORY\" env variable not set")
		} else {
			return repository
		}
	}

	return GetRepositoryName()
}

// GetImageTag returns image tag if defined in IKE_IMAGE_TAG variable or "latest" otherwise.
func GetImageTag() string {
	if imageTag, found := os.LookupEnv("IKE_IMAGE_TAG"); found {
		return imageTag
	}

	return "latest"
}

func SetExternalContainerRegistry() string {
	registry := "default-route-openshift-image-registry." + GetClusterHost()
	if externalRegistry, found := os.LookupEnv("IKE_EXTERNAL_CONTAINER_REGISTRY"); found {
		registry = externalRegistry
	}
	setContainerRegistry(registry)

	return registry
}

func SetInternalContainerRegistry() {
	setContainerRegistry(GetInternalContainerRegistry())
}

func setContainerRegistry(registry string) {
	err := os.Setenv(config.EnvImageRegistry, registry)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
}

func setContainerRepository(namespace string) {
	err := os.Setenv(config.EnvImageRepository, namespace)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
}

// GetInternalContainerRegistry returns the internal address for the container registry.
func GetInternalContainerRegistry() string {
	if internalRegistry, found := os.LookupEnv("IKE_INTERNAL_CONTAINER_REGISTRY"); found {
		return internalRegistry
	}

	return "image-registry.openshift-image-registry.svc:5000"
}
