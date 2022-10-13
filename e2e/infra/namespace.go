package infra

import (
	"os"
	"strconv"

	"github.com/maistra/istio-workspace/pkg/naming"
	testshell "github.com/maistra/istio-workspace/test/shell"
)

// ChangeNamespace switch to different namespace - so we also test -n parameter of $ ike.
// That only works for oc cli, as kubectl by default uses `default` namespace.
func ChangeNamespace(namespace string) {
	if RunsOnOpenshift {
		<-testshell.Execute("oc project " + namespace).Done()
	}
}

func GenerateNamespaceName() string {
	return "ike-tests-" + naming.GenerateString(16)
}

func CleanupNamespace(namespace string, wait bool) {
	if keepStr, found := os.LookupEnv("IKE_E2E_KEEP_NS"); found {
		keep, _ := strconv.ParseBool(keepStr)
		if keep {
			return
		}
	}
	CleanupTestScenario(namespace)
	<-testshell.Execute("kubectl delete namespace " + namespace + " --wait=" + strconv.FormatBool(wait)).Done()
}

// DumpEnvironmentDebugInfo prints tons of noise about the cluster state when test fails.
func DumpEnvironmentDebugInfo(namespace, dir string) {
	GetEvents(namespace)
	DumpTelepresenceLog(dir)
}
