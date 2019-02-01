package main

import (
	"github.com/aslakknutsen/istio-workspace/cmd/ike/cmd"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func main() {
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(logf.ZapLogger(true))

	cmd.Execute()
}
