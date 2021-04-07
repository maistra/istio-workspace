package infra

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/maistra/istio-workspace/test/shell"
)

var RunsAgainstOpenshift = func() bool {
	cmdGetDefaultServices := shell.Execute("kubectl get services -o=custom-columns='SERVICES:metadata.name' --no-headers -n default")
	<-cmdGetDefaultServices.Done()
	defaultServices := strings.Join(cmdGetDefaultServices.Status().Stdout, "")

	return strings.Contains(defaultServices, "openshift")
}()

// UpdateSecurityConstraintsFor applies anyuid and privileged constraints to a given namespace.
func UpdateSecurityConstraintsFor(namespace string) {
	shell.ExecuteAll(
		"oc adm policy add-scc-to-user anyuid -z default -n "+namespace,
		"oc adm policy add-scc-to-user privileged -z default -n "+namespace)
}

func EnablePullingImages(namespace string) {
	<-shell.Execute("oc policy add-role-to-user system:image-puller system:serviceaccount:" + namespace + ":default -n " + GetRepositoryName()).Done()
	<-shell.Execute("oc policy add-role-to-user system:image-puller system:serviceaccount:" + namespace + ":istio-workspace -n " + GetRepositoryName()).Done()
}

var (
	user = "admin"
	pwd  = "admin"
)

func LoginAsTestPowerUser() {
	if ikeUser, found := os.LookupEnv("IKE_CLUSTER_USER"); found {
		user = ikeUser
	}

	if ikePwd, found := os.LookupEnv("IKE_CLUSTER_PWD"); found {
		pwd = ikePwd
	}

	srv := ""
	if server, found := os.LookupEnv("IKE_CLUSTER_ADDRESS"); found {
		srv = server
	}

	<-shell.ExecuteInDir(".", "oc", "login", srv, "-u", user, "-p", pwd, "--insecure-skip-tls-verify=true").Done()
}

// GetEvents returns all events which occurred for a given namespace.
func GetEvents(ns string) {
	state := shell.Execute("kubectl get events -n " + ns)
	<-state.Done()
}

// DumpTelepresenceLog dumps telepresence log if exists.
func DumpTelepresenceLog(dir string) {
	fh, err := os.Open(dir + string(os.PathSeparator) + "telepresence.log")
	if err != nil {
		fmt.Println(err)

		return
	}

	_, err = io.Copy(os.Stdout, fh)
	if err != nil {
		fmt.Println(err)
	}
}

// UsePrebuiltImages returns true if test suite should use images that are built outside of the test execution flow.
func UsePrebuiltImages() bool {
	return os.Getenv("PRE_BUILT_IMAGES") != ""
}

// PrepareEnv sets up a environmental specific things.
func PrepareEnv(namespace string) {
	if RunsAgainstOpenshift {
		UpdateSecurityConstraintsFor(namespace)
		if !UsePrebuiltImages() {
			EnablePullingImages(namespace)
		}
	}
}
