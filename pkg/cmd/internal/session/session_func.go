package internal

import (
	"github.com/maistra/istio-workspace/pkg/internal/session"
	"github.com/maistra/istio-workspace/pkg/telepresence"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Sessions creates a Handler for the given session operation
// session expects that cmd has offline, namespace, route, deployment and session flags defined.
// otherwise it fails.
func Sessions(cmd *cobra.Command) (session.State, session.Options, func(), error) {
	var sessionHandler session.Handler = session.Offline

	if offline, err := cmd.Flags().GetBool("offline"); err == nil && !offline {
		sessionHandler = session.CreateOrJoinHandler
	}

	options, err := ToOptions(cmd.Flags())
	if err != nil {
		return session.State{}, options, nil, err
	}

	state, f, err := sessionHandler(options)

	return state, options, f, err
}

// RemoveSessions creates a Handler for the given session operation for removing a session
// session expects that cmd has offline and session flags defined.
// otherwise it fails.
func RemoveSessions(cmd *cobra.Command) (session.State, func(), error) {
	options, err := ToRemoveOptions(cmd.Flags())
	if err != nil {
		return session.State{}, nil, err
	}
	return session.RemoveHandler(options)
}

const telepresenceStrategy = "telepresence"

// ToOptions converts between FlagSet to a Handler Options.
func ToOptions(flags *pflag.FlagSet) (session.Options, error) {
	strategy := telepresenceStrategy
	strategyArgs := map[string]string{}

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

	i, _ := flags.GetString("image") // ignore error, not a required argument
	if i != "" {
		strategy = "prepared-image"
		strategyArgs["image"] = i
	}

	if strategy == telepresenceStrategy {
		if strategyArgs["version"], err = telepresence.GetVersion(); err != nil {
			return session.Options{}, err
		}
	}

	return session.Options{
		NamespaceName:  n,
		DeploymentName: d,
		SessionName:    s,
		RouteExp:       r,
		Strategy:       strategy,
		StrategyArgs:   strategyArgs,
	}, nil
}

// ToRemoveOptions converts between FlagSet to a Handler Options.
func ToRemoveOptions(flags *pflag.FlagSet) (session.Options, error) {
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

	return session.Options{
		NamespaceName:  n,
		DeploymentName: d,
		SessionName:    s,
	}, nil
}
