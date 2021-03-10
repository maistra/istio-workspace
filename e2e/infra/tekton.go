package infra

import (
	"strings"

	"github.com/maistra/istio-workspace/test/shell"
)

// TaskIsDone checks if given task has succeeded.
func TaskIsDone(ns, taskName string) func() bool {
	return func() bool {
		taskRunStatus := shell.ExecuteInDir(".", "kubectl", "get", "taskruns", taskName, "-n", ns, "-o", "jsonpath='{.status.conditions[?(.type==\"Succeeded\")].reason")
		<-taskRunStatus.Done()
		return strings.Contains(strings.Join(taskRunStatus.Status().Stdout, ""), "Succeeded")
	}
}

// TaskResult returns value of given result variable for defined Task.
func TaskResult(ns, taskName, key string) string {
	taskResultStatus := shell.ExecuteInDir(".", "kubectl", "get", "taskruns", taskName, "-n", ns, "-o", "jsonpath='{.status.taskResults[?(.name==\""+key+"\")].value")
	<-taskResultStatus.Done()
	return strings.Join(taskResultStatus.Status().Stdout, "")
}
