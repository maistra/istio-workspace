package cmd

import (
	"fmt"
	"strings"

	"github.com/maistra/istio-workspace/cmd/ike/config"

	"github.com/spf13/cobra"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("cmd")

// NewRootCmd creates instance of root "ike" Cobra Command with flags and execution logic defined
func NewRootCmd() *cobra.Command {
	var configFile string

	rootCmd := &cobra.Command{
		Use:   "ike",
		Short: "ike lets you safely develop and test on prod without a sweat",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			return config.SetupConfigSources(configFile, cmd.Flag("config").Changed)
		},
		RunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			shouldPrintVersion, _ := cmd.Flags().GetBool("version")
			if shouldPrintVersion {
				printVersion()
			} else {
				fmt.Print(cmd.UsageString())
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().
		StringVarP(&configFile, "config", "c", ".ike.config.yaml",
			fmt.Sprintf("config file (supported formats: %s)", strings.Join(config.SupportedExtensions(), ", ")))
	rootCmd.Flags().Bool("version", false, "prints the version number of ike cli")
	return rootCmd
}
