package cmd

import (
	"fmt"
	"strings"

	"github.com/maistra/istio-workspace/cmd/ike/config"

	"github.com/spf13/cobra"
)

// NewRootCmd creates instance of root "ike" Cobra Command with flags and execution logic defined
func NewRootCmd() *cobra.Command {
	var configFile string

	rootCmd := &cobra.Command{
		Use:   "ike",
		Short: "ike lets you safely develop and test on prod without a sweat",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			return config.SetupConfigSources(configFile, cmd.Flag("config").Changed)
		},
	}

	rootCmd.PersistentFlags().
		StringVarP(&configFile, "config", "c", ".ike.config.yaml",
			fmt.Sprintf("config file (supported formats: %s)", strings.Join(config.SupportedExtensions(), ", ")))

	return rootCmd
}
