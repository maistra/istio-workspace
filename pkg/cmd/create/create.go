package create

import (
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/maistra/istio-workspace/pkg/cmd/config"
	internal "github.com/maistra/istio-workspace/pkg/cmd/internal/session"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/spf13/cobra"
)

var logger = func() logr.Logger {
	return log.Log.WithValues("type", "create")
}

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
			state, _, _, err := internal.Sessions(cmd)
			if outputJSON, _ := cmd.Flags().GetBool("json"); outputJSON {
				b, _ := json.MarshalIndent(&state.Session, "", "  ")
				fmt.Println(string(b))
			}
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
		logger().Error(err, "failed while trying to hide a flag")
	}
	createCmd.Flags().Bool("json", false, "return result in json")
	createCmd.Flag("json").Annotations["silent"] = []string{"true"}

	createCmd.Flags().VisitAll(config.BindFullyQualifiedFlag(createCmd))

	_ = createCmd.MarkFlagRequired("deployment")
	_ = createCmd.MarkFlagRequired("image")

	return createCmd
}
