package infra

import (
	"github.com/maistra/istio-workspace/pkg/shell"
)

// CreateNewApp creates new project with a given name, deploys simple datawire/hello-world app and exposes route to
// it service
func CreateNewApp(name string) {
	<-shell.Execute("oc login -u developer").Done()

	<-shell.Execute("oc new-project " + name).Done()

	UpdateSecurityConstraintsFor(name)

	<-shell.ExecuteInDir(".",
		"oc", "new-app",
		"--docker-image", "datawire/hello-world",
		"--name", name,
		"--allow-missing-images",
	).Done()
	<-shell.Execute("oc expose svc/" + name).Done()
	<-shell.Execute("oc status").Done()
}

// UpdateSecurityConstraintsFor applies anyuid and privileged constraints to a given namespave
func UpdateSecurityConstraintsFor(namespace string) {
	<-shell.Execute("oc login -u system:admin").Done()
	<-shell.Execute("oc adm policy add-scc-to-user anyuid -z default -n " + namespace).Done()
	<-shell.Execute("oc adm policy add-scc-to-user privileged -z default -n" + namespace).Done()
}
