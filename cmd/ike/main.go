package main

import (
	"math/rand"
	"time"

	"github.com/maistra/istio-workspace/cmd/ike/cmd"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func main() {
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(logf.ZapLogger(false))

	// Setting random seed e.g. for session name generator
	rand.Seed(time.Now().UTC().UnixNano())

	rootCmd := cmd.NewRootCmd()
	rootCmd.AddCommand(cmd.NewVersionCmd(),
		cmd.NewDevelopCmd(),
		cmd.NewWatchCmd(),
		cmd.NewServeCmd(),
		cmd.NewInstallCmd(),
	)
	_ = rootCmd.Execute()
}
