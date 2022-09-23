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
			"-n", ns, "-o", "jsonpath='{.items[*].kind}'")
		<-countCmd.Done()
		if countCmd.Status().Error != nil {
			fmt.Println(countCmd.Status().Error.Error())

			return false
		}
		const emptyCmdStdOut = "''"

		if countCmd.Status().Stdout[0] == emptyCmdStdOut {
			return false
		}
		count := len(strings.Split(countCmd.Status().Stdout[0], " "))
		// if nothing is written to stdout it's an error of none found
		if count == 0 {
			return false
		}
		deploymentCmd := shell.ExecuteInDir(".",
			"kubectl", "get", deploymentType,
			"-n", ns,
			"-o", "jsonpath={.items[?(@.status.replicas == @.status.readyReplicas)].metadata.name}")
		<-deploymentCmd.Done()

		// returning nothing at this point means no deployments are in ready state, but some should be
		if len(deploymentCmd.Status().Stdout) > 0 {
			if deploymentCmd.Status().Stdout[0] == emptyCmdStdOut {
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

// MatchResourceCount eventually matcher matching count of resources.
func MatchResourceCount(count int, getCount func() int) func() bool {
	return func() bool {
		return count == getCount()
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

// GetResourceCountFunc wraps GetResourceCount for to be called repeatedly.
func GetResourceCountFunc(resource, ns string) func() int {
	return func() int {
		return GetResourceCount(resource, ns)
	}
}

// GetResourceCount returns the number of "resource"s in the given namespace.
func GetResourceCount(resource, ns string) int {
	cmd := shell.ExecuteInDir(".",
		"kubectl", "get", resource,
		"-n", ns)
	<-cmd.Done()
	if len(cmd.Status().Stdout) == 0 {
		return 0
	}

	return len(cmd.Status().Stdout) - 1
}

type conditionStruct struct {
	Reason string `json:"reason,omitempty"`
	Status string `json:"status,omitempty"`
}

func isPodInStatus(pod, ns, conditionType string) bool {
	podStatus := shell.ExecuteInDir(".",
		"kubectl", "get",
		"pod", pod,
		"-n", ns,
		"-o", `jsonpath-as-json={.status.conditions[?(@.type=="`+conditionType+`")]}`,
	)
	<-podStatus.Done()

	jsonBody := strings.Join(podStatus.Status().Stdout, "")

	var conditions []conditionStruct
	err := json.Unmarshal([]byte(jsonBody), &conditions)
	if err != nil {
		fmt.Println(err)

		return false
	}

	if len(conditions) == 0 {
		return false
	}

	condition := conditions[0]
	status, err := strconv.ParseBool(condition.Status)
	if err != nil {
		return false
	}
	if !status && strings.EqualFold(condition.Reason, "podcompleted") {
		return true
	}

	return status
}
