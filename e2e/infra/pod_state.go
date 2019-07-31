package infra

import (
	"fmt"
	"strings"

	"github.com/maistra/istio-workspace/test/shell"
)

// AllPodsReady checks whether all the pods (and their containers) in the given namespace are in Ready state
func AllPodsReady(ns string) func() bool {
	return func() bool {
		podsCmd := shell.ExecuteInDir(".",
			"oc", "get", "pod",
			"-n", ns,
			"-o", "jsonpath={.items[*].metadata.name}")
		<-podsCmd.Done()
		pods := strings.Split(strings.Trim(fmt.Sprintf("%s", podsCmd.Status().Stdout), "[]"), " ")
		for _, pod := range pods {
			if !isPodReady(pod, ns) {
				return false
			}
		}
		return true
	}
}

func isPodReady(pod, ns string) bool {
	podStatus := shell.ExecuteInDir(".",
		"oc", "get",
		"pod", pod,
		"-n", ns,
		"-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`,
	)
	<-podStatus.Done()
	return strings.Trim(fmt.Sprintf("%s", podStatus.Status().Stdout), "[]") == "True"
}
