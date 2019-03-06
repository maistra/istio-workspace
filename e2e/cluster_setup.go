package e2e

import "github.com/aslakknutsen/istio-workspace/cmd/ike/cmd"

func CreateNewApp(name string) {
	<-cmd.Execute("oc", "new-project", name).Done()

	UpdateSecurityConstraintsFor(name)

	<-cmd.Execute("oc", "new-app",
		"--docker-image", "datawire/hello-world",
		"--name", name,
		"--allow-missing-images",
	).Done()
	<-cmd.Execute("oc", "expose", "svc/"+name).Done()
	<-cmd.Execute("oc", "status").Done()
}

func UpdateSecurityConstraintsFor(namespace string) {
	<-cmd.Execute("oc", "login", "-u", "system:admin").Done()
	<-cmd.Execute("oc", "adm", "policy", "add-scc-to-user", "anyuid", "-z", "default", "-n", namespace).Done()
	<-cmd.Execute("oc", "adm", "policy", "add-scc-to-user", "privileged", "-z", "default", "-n", namespace).Done()
	<-cmd.Execute("oc", "login", "-u", "developer").Done()
}
