package internal

import (
	"emperror.dev/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/maistra/istio-workspace/pkg/internal/session"
	"github.com/maistra/istio-workspace/pkg/telepresence"
)

// Sessions creates a Handler for the given session operation.
// It's expected that cmd has offline, namespace, route, deployment and session flags defined.
// Otherwise it fails.
func Sessions(cmd *cobra.Command) (session.State, session.Options, func(), error) {
	var sessionHandler session.Handler = session.Offline
	var client *session.Client

	options, err := ToOptions(cmd.Annotations, cmd.Flags())
	if err != nil {
		return session.State{}, options, nil, err
	}

	if offline, e := cmd.Flags().GetBool("offline"); e == nil && !offline {
		sessionHandler = session.CreateOrJoinHandler
		c, e2 := session.DefaultClient(options.NamespaceName)
		if e2 != nil {
			return session.State{}, options, func() {}, errors.WrapIf(e2, "failed to get default client")
		}
		client = c
	}

	state, f, err := sessionHandler(options, client)

	return state, options, f, err
}

// RemoveSessions creates a Handler for the given session operation for removing a session
// session expects that cmd has offline and session flags defined.
// otherwise it fails.
func RemoveSessions(cmd *cobra.Command) (session.State, func(), error) {
	options, err := ToRemoveOptions(cmd.Flags())
	if err != nil {
		return session.State{}, nil, errors.WrapIf(err, "failed to create options")
	}
	client, err := session.DefaultClient(options.NamespaceName)
	if err != nil {
		return session.State{}, nil, errors.WrapIf(err, "failed to get default client")
	}
	handler, f := session.RemoveHandler(options, client)

	return handler, f, nil
}

const (
	// AnnotationRevert is the name of the command annotation that is used to control the Revert flag.
	AnnotationRevert     = "revert"
	telepresenceStrategy = "telepresence"
)

// ToOptions converts between FlagSet to a Handler Options.
func ToOptions(annotations map[string]string, flags *pflag.FlagSet) (session.Options, error) {
	strategy := telepresenceStrategy
	strategyArgs := map[string]string{}

	n, err := flags.GetString("namespace")
	if err != nil {
		return session.Options{}, errors.Wrap(err, "failed obtaining namespace flag")
	}

	d, err := flags.GetString("deployment")
	if err != nil {
		return session.Options{}, errors.Wrap(err, "failed obtaining deployment flag")
	}

	s, err := flags.GetString("session")
	if err != nil {
		return session.Options{}, errors.Wrap(err, "failed obtaining session flag")
	}

	r, err := flags.GetString("route")
	if err != nil {
		return session.Options{}, errors.Wrap(err, "failed obtaining route flag")
	}

	i, _ := flags.GetString("image") // ignore error, not a required argument
	if i != "" {
		strategy = "prepared-image"
		strategyArgs["image"] = i
	}

	if strategy == telepresenceStrategy {
		if strategyArgs["version"], err = telepresence.GetVersion(); err != nil {
			return session.Options{}, errors.Wrap(err, "failed obtaining telepresence version")
		}
	}
	revert := false
	if val, found := annotations[AnnotationRevert]; found && val == "true" {
		revert = true
	}

	return session.Options{
		Revert:         revert,
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
		return session.Options{}, errors.Wrap(err, "failed obtaining namespace flag")
	}

	d, err := flags.GetString("deployment")
	if err != nil {
		return session.Options{}, errors.Wrap(err, "failed obtaining deployment flag")
	}

	s, err := flags.GetString("session")
	if err != nil {
		return session.Options{}, errors.Wrap(err, "failed obtaining session flag")
	}

	return session.Options{
		NamespaceName:  n,
		DeploymentName: d,
		SessionName:    s,
	}, nil
}
