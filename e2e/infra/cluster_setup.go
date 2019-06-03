package infra

import "github.com/maistra/istio-workspace/cmd/ike/cmd"

// CreateNewApp creates new project with a given name, deploys simple datawire/hello-world app and exposes route to
// it service
func CreateNewApp(name string) {
	<-cmd.Execute("oc login -u developer").Done()

	<-cmd.Execute("oc new-project " + name).Done()

	UpdateSecurityConstraintsFor(name)

	<-cmd.ExecuteInDir(".",
		"oc", "new-app",
		"--docker-image", "datawire/hello-world",
		"--name", name,
		"--allow-missing-images",
	).Done()
	<-cmd.Execute("oc expose svc/" + name).Done()
	<-cmd.Execute("oc status").Done()
}

// UpdateSecurityConstraintsFor applies anyuid and privileged constraints to a given namespave
func UpdateSecurityConstraintsFor(namespace string) {
	<-cmd.Execute("oc login -u system:admin").Done()
	<-cmd.Execute("oc adm policy add-scc-to-user anyuid -z default -n " + namespace).Done()
	<-cmd.Execute("oc adm policy add-scc-to-user privileged -z default -n" + namespace).Done()
}
