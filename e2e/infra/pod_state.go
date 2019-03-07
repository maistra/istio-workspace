package infra

import (
	"fmt"

	"github.com/aslakknutsen/istio-workspace/cmd/ike/cmd"
)

func PodStatus(namespace, label, state string) func() (string, error) {
	return func() (string, error) {
		ocGetPods := cmd.Execute("oc", "get", "pods",
			"-n", namespace,
			"-l", label,
			"--field-selector", "status.phase=="+state,
		)
		<-ocGetPods.Done()
		return fmt.Sprintf("%v", ocGetPods.Status().Stdout), nil
	}
}

func PodCompletedStatus(namespace, label string) func() (string, error) {
	return func() (string, error) {
		ocGetPods := cmd.Execute("oc", "get", "pods",
			"-n", namespace,
			"-l", label,
			"-o", "go-template='{{range .items}}{{range .status.containerStatuses}}{{.state.terminated.reason}}{{end}}{{end}}'",
		)
		<-ocGetPods.Done()
		return fmt.Sprintf("%v", ocGetPods.Status().Stdout), nil
	}
}
