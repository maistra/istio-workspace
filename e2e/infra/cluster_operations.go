package infra

import (
	"os"
	"strings"

	"github.com/maistra/istio-workspace/test/shell"
)

// UpdateSecurityConstraintsFor applies anyuid and privileged constraints to a given namespace
func UpdateSecurityConstraintsFor(namespace string) {
	LoginAsTestPowerUser()
	shell.ExecuteAll(
		"oc adm policy add-scc-to-user anyuid -z default -n "+namespace,
		"oc adm policy add-scc-to-user privileged -z default -n "+namespace)
}

func LoginAsTestPowerUser() {
	user := "admin" //nolint[:goconst]
	pwd := "admin"  //nolint[:goconst]
	if ikeUser, found := os.LookupEnv("IKE_CLUSTER_ADMIN"); found {
		user = ikeUser
	}

	if ikePwd, found := os.LookupEnv("IKE_CLUSTER_ADMIN_PWD"); found {
		pwd = ikePwd
	}

	<-shell.Execute("oc login -u " + user + " -p " + pwd).Done()
}

func ClientVersion() int {
	version := shell.Execute("oc version")
	<-version.Done()
	v := strings.Join(version.Status().Stdout, " ")
	if strings.Contains(v, "GitVersion:\"v4.") {
		return 4
	}
	return 3
}
