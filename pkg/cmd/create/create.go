package create

import (
	"github.com/maistra/istio-workspace/pkg/cmd/config"
	internal "github.com/maistra/istio-workspace/pkg/cmd/internal/session"
	"github.com/maistra/istio-workspace/pkg/log"

	"github.com/spf13/cobra"
)

var logger = log.CreateOperatorAwareLogger("cmd").WithValues("type", "create")

// NewCmd creates instance of "create" Cobra Command with flags and execution logic defined.
func NewCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:          "create",
		Short:        "Creates a new Session",
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return config.SyncFullyQualifiedFlags(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _, _, err := internal.Sessions(cmd)
			return err
		},
	}

	createCmd.Flags().StringP("deployment", "d", "", "name of the deployment or deployment config")
	createCmd.Flags().StringP("session", "s", "", "create or join an existing session")
	createCmd.Flags().StringP("image", "i", "", "create a prepared session with the given image")
	createCmd.Flags().StringP("route", "", "", "specifies traffic route options in the format of type:name=value. "+
		"Defaults to X-Workspace-Route header with current session name value")
	createCmd.Flags().StringP("namespace", "n", "", "target namespace to develop against "+
		"(defaults to default for the current context)")
	createCmd.Flags().Bool("offline", false, "avoid calling external sources")
	if err := createCmd.Flags().MarkHidden("offline"); err != nil {
		logger.Error(err, "failed while trying to hide a flag")
	}

	createCmd.Flags().VisitAll(config.BindFullyQualifiedFlag(createCmd))

	_ = createCmd.MarkFlagRequired("deployment")
	_ = createCmd.MarkFlagRequired("image")

	return createCmd
}
