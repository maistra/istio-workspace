package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/maistra/istio-workspace/pkg/cmd/completion"

	"github.com/maistra/istio-workspace/pkg/cmd/version"

	"github.com/maistra/istio-workspace/pkg/cmd/format"

	"github.com/maistra/istio-workspace/pkg/cmd/config"

	v "github.com/maistra/istio-workspace/version"

	"github.com/spf13/cobra"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("cmd").WithValues("type", "root")

// NewCmd creates instance of root "ike" Cobra Command with flags and execution logic defined
func NewCmd() *cobra.Command {
	var configFile string
	releaseInfo := make(chan string, 1)
	timeout := make(chan bool, 1)

	rootCmd := &cobra.Command{
		Use:                    "ike",
		Short:                  "ike lets you safely develop and test on prod without a sweat",
		BashCompletionFunction: completion.BashCompletionFunc,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			if v.Released() {
				go func() {
					time.Sleep(2 * time.Second)
					timeout <- true
				}()
				go func() {
					latestRelease, _ := version.LatestRelease()
					if !version.IsLatestRelease(latestRelease) {
						releaseInfo <- fmt.Sprintf("WARN: you are using %s which is not the latest release (newest is %s). "+
							"You can update it using:\n$ curl -sL http://git.io/get-ike | bash\n", v.Version, latestRelease)
					} else {
						releaseInfo <- ""
					}
				}()
			}
			return config.SetupConfigSources(configFile, cmd.Flag("config").Changed)
		},
		RunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			shouldPrintVersion, _ := cmd.Flags().GetBool("version")
			if shouldPrintVersion {
				version.PrintVersion()
			} else {
				fmt.Print(cmd.UsageString())
			}
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if v.Released() {
				select {
				case release := <-releaseInfo:
					log.Info(release)
				case <-timeout:
					// do nothing, just timeout
				}
			}
			close(releaseInfo)
			close(timeout)
			return nil
		},
	}

	rootCmd.PersistentFlags().
		StringVarP(&configFile, "config", "c", ".ike.config.yaml",
			fmt.Sprintf("config file (supported formats: %s)", strings.Join(config.SupportedExtensions(), ", ")))
	rootCmd.Flags().Bool("version", false, "prints the version number of ike cli")
	rootCmd.PersistentFlags().String("help-format", "standard", "prints help in asciidoc table")
	if err := rootCmd.PersistentFlags().MarkHidden("help-format"); err != nil {
		log.Error(err, "failed while trying to hide a flag")
	}

	format.EnhanceHelper(rootCmd)
	format.RegisterTemplateFuncs()

	return rootCmd
}
