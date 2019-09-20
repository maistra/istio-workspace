package create

import (
	"github.com/maistra/istio-workspace/pkg/cmd/config"
	internal "github.com/maistra/istio-workspace/pkg/cmd/internal/session"

	"github.com/spf13/cobra"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("cmd").WithValues("type", "create")

// NewCmd creates instance of "create" Cobra Command with flags and execution logic defined
func NewCmd() *cobra.Command {
	developCmd := &cobra.Command{
		Use:          "create",
		Short:        "Creates a new Session",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			_, _, err := internal.Sessions(cmd)
			return err
		},
	}

	developCmd.Flags().StringP("deployment", "d", "", "name of the deployment or deployment config")
	developCmd.Flags().StringP("session", "s", "", "create or join an existing session")
	developCmd.Flags().StringP("image", "i", "", "create a prepared session with the given image")
	developCmd.Flags().StringP("route", "", "", "specifies traffic route options in the format of type:name=value. "+
		"Defaults to X-Workspace-Route header with current session name value")
	developCmd.Flags().StringP("namespace", "n", "", "target namespace to develop against "+
		"(defaults to default for the current context)")
	developCmd.Flags().Bool("offline", false, "avoid calling external sources")
	if err := developCmd.Flags().MarkHidden("offline"); err != nil {
		log.Error(err, "failed while trying to hide a flag")
	}

	developCmd.Flags().VisitAll(config.BindFullyQualifiedFlag(developCmd))

	_ = developCmd.MarkFlagRequired("deployment")
	_ = developCmd.MarkFlagRequired("image")

	return developCmd
}
