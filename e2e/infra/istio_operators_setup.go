package infra

import (
	"fmt"
	"os"

	"github.com/maistra/istio-workspace/test/shell"

	"github.com/onsi/gomega"
)

// LoadIstio calls make load-istio target and waits until operator sets up the mesh
func LoadIstio() {
	projectDir := shell.GetProjectDir()
	<-shell.ExecuteInDir(projectDir, "make", "load-istio").Done()
}

// BuildOperator builds istio-workspace operator and pushes it to specified registry
func BuildOperator() (registry string) {
	projectDir := shell.GetProjectDir()
	_, registry = setDockerEnvForOperatorBuild()
	LoginAsTestPowerUser()
	<-shell.ExecuteInDir(".", "bash", "-c", "docker login -u $(oc whoami) -p $(oc whoami -t) "+registry).Done()
	<-shell.ExecuteInDir(projectDir, "make", "docker-build").Done()
	return
}

// PushOperatorImage deploys istio-workspace operator into specified namespace
func PushOperatorImage(namespace string) {
	projectDir := shell.GetProjectDir()
	gomega.Expect(projectDir).To(gomega.Not(gomega.BeEmpty()))
	LoginAsTestPowerUser()

	setDockerEnvForLocalOperatorBuild(namespace)
	_ = os.Setenv("IKE_IMAGE_NAME", "istio-workspace")
	<-shell.ExecuteInDir(".", "bash", "-c", "docker tag $IKE_DOCKER_REGISTRY/istio-workspace-operator/$IKE_IMAGE_NAME:$IKE_IMAGE_TAG $IKE_DOCKER_REGISTRY/"+namespace+"/$IKE_IMAGE_NAME:$IKE_IMAGE_TAG").Done() //nolint[:lll]
	<-shell.ExecuteInDir(".", "bash", "-c", "docker push $IKE_DOCKER_REGISTRY/"+namespace+"/$IKE_IMAGE_NAME:$IKE_IMAGE_TAG").Done()

	setDockerEnvForLocalOperatorDeploy(namespace)
}

func InstallLocalOperator(namespace string) {
	<-shell.Execute("ike install-operator -l -n " + namespace).Done()
}

func setOperatorNamespace() (namespace string) {
	operatorNS := "istio-workspace-operator"

	err := os.Setenv("OPERATOR_NAMESPACE", operatorNS)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	setDockerRepository(operatorNS)
	return operatorNS
}

func setLocalOperatorNamespace(namespace string) {
	err := os.Setenv("OPERATOR_NAMESPACE", namespace)
	gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	setDockerRepository(namespace)
}

func GetClusterHost() string {
	if host, found := os.LookupEnv("IKE_CLUSTER_HOST"); found {
		return host
	}
	return "127.0.0.1.nip.io"
}

func GetIstioNamespace() string {
	if istioNs, found := os.LookupEnv("ISTIO_NS"); found {
		return istioNs
	}
	return "istio-system"
}

func GetIstioIngressHostname() string {
	cmd := shell.ExecuteInDir(".", "bash", "-c", fmt.Sprintf("oc get svc istio-ingressgateway -n %v -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'", GetIstioNamespace()))
	<-cmd.Done()
	if cmd.Status().Exit == 0 && len(cmd.Status().Stdout) > 0 {
		return "http://" + cmd.Status().Stdout[0]
	}
	cmd = shell.ExecuteInDir(".", "bash", "-c", fmt.Sprintf("oc get svc istio-ingressgateway -n %v -o jsonpath='{.spec.clusterIP}'", GetIstioNamespace()))
	<-cmd.Done()
	if cmd.Status().Exit == 0 && len(cmd.Status().Stdout) > 0 {
		return "http://" + cmd.Status().Stdout[0]
	}
	return "http://istio-ingressgateway-" + GetIstioNamespace() + "." + GetClusterHost()
}
