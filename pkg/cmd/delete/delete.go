package delete

import (
	"github.com/maistra/istio-workspace/pkg/cmd/config"
	internal "github.com/maistra/istio-workspace/pkg/cmd/internal/session"

	"github.com/spf13/cobra"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("cmd").WithValues("type", "delete")

// NewCmd creates instance of "create" Cobra Command with flags and execution logic defined
func NewCmd() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:          "delete",
		Short:        "Deletes an existing Session",
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			return config.SyncFullyQualifiedFlags(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			_, remove, err := internal.RemoveSessions(cmd)
			if err == nil {
				remove()
			}
			return err
		},
	}

	deleteCmd.Flags().StringP("deployment", "d", "", "name of the deployment or deployment config")
	deleteCmd.Flags().StringP("session", "s", "", "create or join an existing session")
	deleteCmd.Flags().StringP("namespace", "n", "", "target namespace to develop against "+
		"(defaults to default for the current context)")
	deleteCmd.Flags().Bool("offline", false, "avoid calling external sources")
	if err := deleteCmd.Flags().MarkHidden("offline"); err != nil {
		log.Error(err, "failed while trying to hide a flag")
	}

	deleteCmd.Flags().VisitAll(config.BindFullyQualifiedFlag(deleteCmd))

	_ = deleteCmd.MarkFlagRequired("deployment")
	_ = deleteCmd.MarkFlagRequired("session")

	return deleteCmd
}
