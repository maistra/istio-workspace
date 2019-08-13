package infra

import (
	"github.com/maistra/istio-workspace/test/shell"
	"os"
)

// CreateNewApp creates new project with a given name, deploys simple datawire/hello-world app and exposes route to
// it service
func CreateNewApp(name string) {
	shell.ExecuteAll("oc login -u developer", "oc new-project "+name)

	UpdateSecurityConstraintsFor(name)

	<-shell.ExecuteInDir(".",
		"oc", "new-app",
		"--docker-image", "datawire/hello-world",
		"--name", name,
		"--allow-missing-images",
	).Done()
	shell.ExecuteAll("oc expose svc/"+name, "oc status")
}

// UpdateSecurityConstraintsFor applies anyuid and privileged constraints to a given namespace
func UpdateSecurityConstraintsFor(namespace string) {
	LoginAsAdminUser()
	shell.ExecuteAll(
		"oc adm policy add-scc-to-user anyuid -z default -n "+namespace,
		"oc adm policy add-scc-to-user privileged -z default -n"+namespace)
}


func LoginAsAdminUser() {
	user := "admin"
	pwd := "admin"
	if ikeUser, found := os.LookupEnv("IKE_CLUSTER_ADMIN"); found {
		user = ikeUser
	}

	if ikePwd, found := os.LookupEnv("IKE_CLUSTER_ADMIN_PWD"); found {
		pwd = ikePwd
	}

	<-shell.Execute("oc login -u " + user + " -p " + pwd).Done()
}
