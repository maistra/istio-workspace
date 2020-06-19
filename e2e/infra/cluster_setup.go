package infra

import (
	"fmt"
	"io"
	"os"

	"github.com/maistra/istio-workspace/test/shell"
)

// UpdateSecurityConstraintsFor applies anyuid and privileged constraints to a given namespace.
func UpdateSecurityConstraintsFor(namespace string) {
	shell.ExecuteAll(
		"oc adm policy add-scc-to-user anyuid -z default -n "+namespace,
		"oc adm policy add-scc-to-user privileged -z default -n "+namespace)
}

func EnablePullingImages(namespace string) {
	<-shell.Execute("oc policy add-role-to-user system:image-puller system:serviceaccount:" + namespace + ":default -n " + ImageRepo).Done()
	<-shell.Execute("oc policy add-role-to-user system:image-puller system:serviceaccount:" + namespace + ":istio-workspace -n " + ImageRepo).Done()
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

	<-shell.ExecuteInDir(".", "bash", "-c", "oc login "+srv+" -u "+user+" -p "+pwd+" --insecure-skip-tls-verify=true").Done()
}

// GetEvents returns all events which occurred for a given namespace.
func GetEvents(ns string) {
	state := shell.Execute("oc get events -n " + ns)
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
