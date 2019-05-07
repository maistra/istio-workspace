package cmd

import (
	"github.com/maistra/istio-workspace/cmd/ike/session"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// session expects that cmd has offline, deployment and session flags defined.
// otherwise it fails
func sessions(cmd *cobra.Command) (session.State, func(), error) {
	var sessionHandler session.Handler = session.Offline

	if offline, err := cmd.Flags().GetBool("offline"); err == nil && !offline {
		sessionHandler = session.CreateOrJoinHandler
	}

	options, err := ToOptions(cmd.Flags())
	if err != nil {
		return session.State{}, nil, err
	}
	return sessionHandler(options)
}

// ToOptions converts between FlagSet to a Handler Options
func ToOptions(flags *pflag.FlagSet) (session.Options, error) {
	n, err := flags.GetString("namespace")
	if err != nil {
		return session.Options{}, err
	}

	d, err := flags.GetString("deployment")
	if err != nil {
		return session.Options{}, err
	}

	s, err := flags.GetString("session")
	if err != nil {
		return session.Options{}, err
	}

	r, err := flags.GetString("route")
	if err != nil {
		return session.Options{}, err
	}

	return session.Options{
		NamespaceName:  n,
		DeploymentName: d,
		SessionName:    s,
		RouteExp:       r,
	}, nil
}
