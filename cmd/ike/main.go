package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/maistra/istio-workspace/pkg/cmd"
	"github.com/maistra/istio-workspace/pkg/cmd/completion"
	"github.com/maistra/istio-workspace/pkg/cmd/develop"
	"github.com/maistra/istio-workspace/pkg/cmd/install"
	"github.com/maistra/istio-workspace/pkg/cmd/serve"
	"github.com/maistra/istio-workspace/pkg/cmd/version"
	"github.com/maistra/istio-workspace/pkg/cmd/watch"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func main() {
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniformed and structured logs.
	// Logs to os.Stderr, where all structured logging should go
	// When running outside of k8s cluster it will use development
	// mode so the log is not in JSON, but plain text format
	logf.SetLogger(logf.ZapLogger(!isRunningInK8sCluster()))

	// Setting random seed e.g. for session name generator
	rand.Seed(time.Now().UTC().UnixNano())

	rootCmd := cmd.NewCmd()
	rootCmd.AddCommand(version.NewCmd(),
		develop.NewCmd(),
		watch.NewCmd(),
		serve.NewCmd(),
		install.NewCmd(),
		completion.NewCmd(),
	)

	cmd.VisitAll(rootCmd, completion.AddFlagCompletion)

	_ = rootCmd.Execute()
}

func isRunningInK8sCluster() bool {
	_, runningInCluster := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	return runningInCluster
}
