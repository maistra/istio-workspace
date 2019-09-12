package main

import (
	"math/rand"
	"time"

	"github.com/maistra/istio-workspace/pkg/log"

	"github.com/maistra/istio-workspace/pkg/cmd"
	"github.com/maistra/istio-workspace/pkg/cmd/completion"
	"github.com/maistra/istio-workspace/pkg/cmd/develop"
	"github.com/maistra/istio-workspace/pkg/cmd/install"
	"github.com/maistra/istio-workspace/pkg/cmd/serve"
	"github.com/maistra/istio-workspace/pkg/cmd/version"
	"github.com/maistra/istio-workspace/pkg/cmd/watch"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func main() {
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniformed and structured logs.
	// Logs to os.Stderr, where all structured logging should go
	// When running outside of k8s cluster it will use development
	// mode so the log is not in JSON, but plain text format
	logger := log.CreateClusterAwareLogger()
	logf.SetLogger(logger)

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

	if err := rootCmd.Execute(); err != nil {
		logger.Error(err, "failed executing command")
	}
}
