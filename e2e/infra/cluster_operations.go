package infra

import (
	"fmt"
	"io"
	"os"
	"strconv"
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

func ClientVersion() int {
	if version, found := os.LookupEnv("IKE_CLUSTER_VERSION"); found {
		result, _ := strconv.ParseInt(version, 0, 0)
		return int(result)
	}
	version := shell.Execute("oc version")
	<-version.Done()
	v := strings.Join(version.Status().Stdout, " ")
	if strings.Contains(v, "Server Version: 4.") {
		return 4
	}
	return 3
}

// GetEvents returns all events which occurred for a given namespace
func GetEvents(ns string) {
	state := shell.Execute("oc get events -n " + ns)
	<-state.Done()
}

// DumpTelepresenceLog dumps telepresence log if exists
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
