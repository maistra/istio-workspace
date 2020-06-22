package infra

import (
	"fmt"
	"strings"

	"github.com/maistra/istio-workspace/test/shell"
)

// AllPodsReady checks whether all the pods (and their containers) in the given namespace are in Ready state.
func AllPodsReady(ns string) func() bool {
	return func() bool {
		pods := GetAllPods(ns)
		for _, pod := range pods {
			if strings.Contains(pod, "-deploy") {
				fmt.Printf("Skipping deploy pod %s\n", pod)
				continue
			}
			if !isPodInStatus(pod, ns, "Ready") {
				// but skip Completed
				return false
			}
		}
		return true
	}
}

// GetAllPods returns names of all pods from a given namespace.
func GetAllPods(ns string) []string {
	podsCmd := shell.ExecuteInDir(".",
		"oc", "get", "pod",
		"-n", ns,
		"-o", "jsonpath={.items[*].metadata.name}")
	<-podsCmd.Done()
	return strings.Split(strings.Trim(fmt.Sprintf("%s", podsCmd.Status().Stdout), "[]"), " ")
}

// StateOf returns state of the pod.
func StateOf(ns, pod string) {
	state := shell.Execute("oc get pod " + pod + " -n " + ns + " -o yaml")
	<-state.Done()
}

// LogsOf returns logs of all containers in the pod.
func LogsOf(ns, pod string) string {
	logs := shell.Execute("oc logs " + pod + " -n " + ns + " --all-containers=true")
	<-logs.Done()
	return fmt.Sprintf("%s", logs.Status().Stdout)
}

func isPodInStatus(pod, ns, status string) bool {
	podStatus := shell.ExecuteInDir(".",
		"oc", "get",
		"pod", pod,
		"-n", ns,
		"-o", `jsonpath={.status.conditions[?(@.type=="`+status+`")].status}`,
	)
	<-podStatus.Done()
	return strings.Trim(fmt.Sprintf("%s", podStatus.Status().Stdout), "[]") == "True"
}
