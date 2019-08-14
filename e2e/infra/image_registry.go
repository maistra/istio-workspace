package infra

import (
	"os"

	"github.com/maistra/istio-workspace/pkg/cmd/config"

	"github.com/onsi/gomega"
)

func setDockerEnvForOperatorBuild() (namespace, registry string) {
	ns := setOperatorNamespace()
	repo := setDockerRegistryExternal()
	return ns, repo
}

func setDockerEnvForLocalOperatorBuild(namespace string) string {
	setLocalOperatorNamespace(namespace)
	repo := setDockerRegistryExternal()
	return repo
}

func setDockerEnvForOperatorDeploy() (namespace, registry string) {
	ns := setOperatorNamespace()
	repo := setDockerRegistryInternal()
	return ns, repo
}

func setDockerEnvForLocalOperatorDeploy(namespace string) string {
	setLocalOperatorNamespace(namespace)
	repo := setDockerRegistryInternal()
	return repo
}

func setDockerRegistryExternal() string {
	var registry string
	switch ClientVersion() {
	case 3:
		registry = "docker-registry-default." + getClusterHost() + ":80"
	case 4:
		registry = "default-route-openshift-image-registry.apps." + getClusterHost()
	}

	setDockerRegistry(registry)
	return registry
}

func setDockerRegistry(registry string) {
	err := os.Setenv(config.EnvDockerRegistry, registry)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
}

func setDockerRepository(namespace string) {
	err := os.Setenv(config.EnvDockerRepository, namespace)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
}

func setDockerRegistryInternal() string {
	var registry string
	switch ClientVersion() {
	case 3:
		registry = "172.30.1.1:5000"
	case 4:
		registry = "image-registry.openshift-image-registry.svc:5000"
	}
	setDockerRegistry(registry)
	return registry
}
