package infra

import (
	"fmt"
	"strings"

	"github.com/maistra/istio-workspace/test/shell"
)

// AllPodsNotInState creates a function which checks if all pods in the given namespace are in desired state
// Returns content of Stderr to determine if there was an error
func AllPodsNotInState(namespace, state string) func() string {
	return func() string {
		ocGetAllPods := shell.ExecuteInDir(".",
			"oc", "get", "pods",
			"-n", namespace,
		)
		<-ocGetAllPods.Done()

		if strings.Contains(fmt.Sprintf("%v", ocGetAllPods.Status().Stderr), "No resources found") {
			return fmt.Sprintf("no pods in any state found in %s namespace", namespace)
		}

		ocGetFilteredPods := shell.ExecuteInDir(".",
			"oc", "get", "pods",
			"-n", namespace,
			"--field-selector", "status.phase!="+state,
		)
		<-ocGetFilteredPods.Done()
		return fmt.Sprintf("%v", ocGetFilteredPods.Status().Stderr)
	}
}
