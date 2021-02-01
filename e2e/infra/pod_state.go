package infra

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/maistra/istio-workspace/test/shell"
)

// AllDeploymentsAndPodsReady checks if both AllDeploymentsReady(Deployment) and AllPodsReady return true.
func AllDeploymentsAndPodsReady(ns string) func() bool {
	return func() bool {
		return AllDeploymentsReady("deployment", ns)() && AllPodsReady(ns)()
	}
}

// AllDeploymentConfigsAndPodsReady checks if both AllDeploymentsReady(DeploymentConfig) and AllPodsReady return true.
func AllDeploymentConfigsAndPodsReady(ns string) func() bool {
	return func() bool {
		return AllDeploymentsReady("deploymentconfig", ns)() && AllPodsReady(ns)()
	}
}

// AllDeploymentsReady checks whether all the deploymentType(deployment or deploymentconfig) in the given namespace have the same replicas and readyReplicas count.
func AllDeploymentsReady(deploymentType, ns string) func() bool {
	return func() bool {
		countCmd := shell.ExecuteInDir(".",
			"kubectl", "get", deploymentType,
			"-n", ns)
		<-countCmd.Done()
		count := len(countCmd.Status().Stdout)
		// if deployments are found then there is a HEADER
		if count > 1 {
			count--
		}
		// if nothing is written to stdout it's an error of none found
		if count == 0 {
			return true
		}
		deploymentCmd := shell.ExecuteInDir(".",
			"kubectl", "get", deploymentType,
			"-n", ns,
			"-o", "jsonpath={.items[?(@.status.replicas == @.status.readyReplicas)].metadata.name}")
		<-deploymentCmd.Done()

		// returning nothing at this point means no deployments are in ready state, but some should be
		if len(deploymentCmd.Status().Stdout) > 0 {
			if deploymentCmd.Status().Stdout[0] == "" {
				return false
			}
			if len(strings.Split(deploymentCmd.Status().Stdout[0], " ")) == count {
				return true
			}
		}

		return false
	}
}

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
		"kubectl", "get", "pod",
		"-n", ns,
		"-o", "jsonpath={.items[*].metadata.name}")
	<-podsCmd.Done()
	if len(podsCmd.Status().Stdout) == 0 {
		return []string{}
	}
	return strings.Split(strings.Trim(fmt.Sprintf("%s", podsCmd.Status().Stdout), "[]"), " ")
}

// StateOf returns state of the pod.
func StateOf(ns, pod string) {
	state := shell.Execute("kubectl get pod " + pod + " -n " + ns + " -o yaml")
	<-state.Done()
}

// LogsOf returns logs of all containers in the pod.
func LogsOf(ns, pod string) {
	logs := shell.Execute("kubectl logs " + pod + " -n " + ns + " --all-containers=true")
	<-logs.Done()
}

func isPodInStatus(pod, ns, conditionType string) bool {
	podStatus := shell.ExecuteInDir(".",
		"kubectl", "get",
		"pod", pod,
		"-n", ns,
		"-o", `jsonpath-as-json={.status.conditions[?(@.type=="`+conditionType+`")]}`,
	)
	<-podStatus.Done()

	conditions := []map[string]string{}
	err := json.Unmarshal([]byte(fmt.Sprintf("%s", podStatus.Status().Stdout)), &conditions)
	if err != nil {
		return false
	}

	if len(conditions) == 0 {
		return false
	}

	condition := conditions[0]
	status, err := strconv.ParseBool(condition["status"])
	if err != nil {
		return false
	}
	if !status && strings.ToLower(condition["reason"]) == "podcompleted" {
		return true
	}
	return status
}
