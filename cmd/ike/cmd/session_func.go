package cmd

import (
	"github.com/aslakknutsen/istio-workspace/cmd/ike/session"
	"github.com/spf13/cobra"
)

// session expects that cmd has offline, deployment and session flags defined.
// otherwise it fails
func sessions(cmd *cobra.Command) (closer func(), err error) {
	var sessionHandler session.Handler = session.Offline

	if offline, err := cmd.Flags().GetBool("offline"); err == nil && !offline {
		sessionHandler = session.CreateOrJoinHandler
	}

	return sessionHandler(cmd)
}
