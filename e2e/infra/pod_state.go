package infra

import (
	"fmt"

	"github.com/maistra/istio-workspace/cmd/ike/cmd"
)

func AllPodsNotInState(namespace, state string) func() string {
	return func() string {
		ocGetPods := cmd.ExecuteInDir(".",
			"oc", "get", "pods",
			"-n", namespace,
			"--field-selector", "status.phase!="+state,
		)
		<-ocGetPods.Done()
		return fmt.Sprintf("%v", ocGetPods.Status().Stderr)
	}
}

func PodStatus(namespace, label, state string) func() string { //nolint[:unused]
	return func() string {
		ocGetPods := cmd.ExecuteInDir(".",
			"oc", "get", "pods",
			"-n", namespace,
			"-l", label,
			"--field-selector", "status.phase=="+state,
		)
		<-ocGetPods.Done()
		return fmt.Sprintf("%v", ocGetPods.Status().Stdout)
	}
}

func PodCompletedStatus(namespace, label string) func() string {
	return func() string {
		ocGetPods := cmd.ExecuteInDir(".",
			"oc", "get", "pods",
			"-n", namespace,
			"-l", label,
			"-o", "go-template='{{range .items}}{{range .status.containerStatuses}}{{.state.terminated.reason}}{{end}}{{end}}'",
		)
		<-ocGetPods.Done()
		return fmt.Sprintf("%v", ocGetPods.Status().Stdout)
	}
}
