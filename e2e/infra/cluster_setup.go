package infra

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/maistra/istio-workspace/test/shell"
)

var RunsOnOpenshift = func() bool {
	cmdGetDefaultServices := shell.Execute("kubectl get services -o=custom-columns='SERVICES:metadata.name' --no-headers -n default")
	<-cmdGetDefaultServices.Done()
	defaultServices := strings.Join(cmdGetDefaultServices.Status().Stdout, "")

	return strings.Contains(defaultServices, "openshift")
}()

// UpdateSecurityConstraintsFor applies anyuid and privileged constraints to a given namespace.
func UpdateSecurityConstraintsFor(namespace string) {
	shell.WaitForSuccess(
		shell.Execute("oc adm policy add-scc-to-user anyuid -z default -n "+namespace),
		shell.Execute("oc adm policy add-scc-to-user privileged -z default -n "+namespace),
	)
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

// PrintEvents returns all events which occurred for a given namespace.
func PrintEvents(ns string) {
	<-shell.Execute("kubectl get events -n " + ns).Done()
}

// DumpTelepresenceLog dumps telepresence log if exists.
func DumpTelepresenceLog(dir string) {
	fh, err := os.Open(filepath.Join(dir, "telepresence.log"))
	if err != nil {
		fmt.Println(err)

		return
	}

	_, err = io.Copy(os.Stdout, fh)
	if err != nil {
		fmt.Println(err)
	}
}

func PrintControllerLogs(ns string) {
	<-shell.Execute("kubectl logs -l app=istio-workspace --all-containers -n " + ns).Done()
}

// UsePrebuiltImages returns true if test suite should use images that are built outside the test execution flow.
func UsePrebuiltImages() bool {
	return os.Getenv("PRE_BUILT_IMAGES") != ""
}

// PrepareEnv sets up a environmental specific things.
func PrepareEnv(namespace string) {
	if RunsOnOpenshift {
		UpdateSecurityConstraintsFor(namespace)
		if !UsePrebuiltImages() {
			EnablePullingImages(namespace)
		}
	}
}
