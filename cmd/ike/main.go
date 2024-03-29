package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/maistra/istio-workspace/pkg/cmd"
	"github.com/maistra/istio-workspace/pkg/cmd/completion"
	"github.com/maistra/istio-workspace/pkg/cmd/create"
	"github.com/maistra/istio-workspace/pkg/cmd/delete"
	"github.com/maistra/istio-workspace/pkg/cmd/develop"
	"github.com/maistra/istio-workspace/pkg/cmd/execute"
	"github.com/maistra/istio-workspace/pkg/cmd/serve"
	"github.com/maistra/istio-workspace/pkg/cmd/version"
	"github.com/maistra/istio-workspace/pkg/hook"
	"github.com/maistra/istio-workspace/pkg/k8s"
	"github.com/maistra/istio-workspace/pkg/log"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniformed and structured logs.
	// Logs to os.Stderr, where all structured logging should go
	// When running outside of k8s cluster it will use development
	// mode so the log is not in JSON, but plain text format
	log.SetLogger(log.CreateOperatorAwareLogger("root"))

	// Setting random seed e.g. for session name generator
	rand.Seed(time.Now().UTC().UnixNano())

	rootCmd := cmd.NewCmd(&k8s.ClusterVerifier{})
	rootCmd.AddCommand(
		version.NewCmd(),
		create.NewCmd(),
		delete.NewCmd(),
		develop.NewCmd(),
		execute.NewCmd(),
		serve.NewCmd(),
		completion.NewCmd(),
	)

	cmd.VisitAll(rootCmd, completion.AddFlagCompletion)

	if err := rootCmd.Execute(); err != nil {
		log.Log.Error(err, "failed executing command")
		hook.Close()
		os.Exit(23)
	}
}
