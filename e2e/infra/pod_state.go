package infra

import (
	"fmt"
	"strings"

	"github.com/maistra/istio-workspace/pkg/shell"
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

// PodStatus creates a function which examines if there are pods in a given namespace with a label and status
// Returns content of Stdout to inspect
func PodStatus(namespace, label, state string) func() string { //nolint[:unused]
	return func() string {
		ocGetPods := shell.ExecuteInDir(".",
			"oc", "get", "pods",
			"-n", namespace,
			"-l", label,
			"--field-selector", "status.phase=="+state,
		)
		<-ocGetPods.Done()
		return fmt.Sprintf("%v", ocGetPods.Status().Stdout)
	}
}

// PodCompletedStatus creates a func which check if there are pods in a given namespace with desired label which
// are in the terminated state.
// Returns content of Stdout for further inspection
func PodCompletedStatus(namespace, label string) func() string {
	return func() string {
		ocGetPods := shell.ExecuteInDir(".",
			"oc", "get", "pods",
			"-n", namespace,
			"-l", label,
			"-o", "go-template='{{range .items}}{{range .status.containerStatuses}}{{.state.terminated.reason}}{{end}}{{end}}'",
		)
		<-ocGetPods.Done()
		return fmt.Sprintf("%v", ocGetPods.Status().Stdout)
	}
}
