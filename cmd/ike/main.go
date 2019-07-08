package main

import (
	"math/rand"
	"time"

	"github.com/maistra/istio-workspace/pkg/cmd/install"
	"github.com/maistra/istio-workspace/pkg/cmd/serve"
	"github.com/maistra/istio-workspace/pkg/cmd/version"

	"github.com/maistra/istio-workspace/pkg/cmd/develop"

	"github.com/maistra/istio-workspace/pkg/cmd/watch"

	"github.com/maistra/istio-workspace/pkg/cmd"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func main() {
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	// Logs to os.Stderr, where all structured logging should go
	logf.SetLogger(logf.ZapLogger(false))

	// Setting random seed e.g. for session name generator
	rand.Seed(time.Now().UTC().UnixNano())

	rootCmd := cmd.NewRootCmd()
	rootCmd.AddCommand(version.NewVersionCmd(),
		develop.NewDevelopCmd(),
		watch.NewWatchCmd(),
		serve.NewServeCmd(),
		install.NewInstallCmd(),
	)
	_ = rootCmd.Execute()
}
