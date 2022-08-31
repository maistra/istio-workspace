package cmd

import (
	"fmt"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"

	"github.com/maistra/istio-workspace/pkg/cmd/completion"
	"github.com/maistra/istio-workspace/pkg/cmd/config"
	"github.com/maistra/istio-workspace/pkg/cmd/format"
	"github.com/maistra/istio-workspace/pkg/cmd/version"
	"github.com/maistra/istio-workspace/pkg/hook"
	"github.com/maistra/istio-workspace/pkg/log"
	v "github.com/maistra/istio-workspace/version"
)

var logger = func() logr.Logger {
	return log.Log.WithValues("type", "root")
}

// NewCmd creates instance of root "ike" Cobra Command with flags and execution logic defined.
func NewCmd() *cobra.Command {
	var configFile string
	releaseInfo := make(chan string, 1)

	released := v.Released()
	rootCmd := &cobra.Command{
		SilenceErrors: true,
		Use:           "ike",
		Short: "ike lets you safely develop and test on production without a sweat!\n\n" +
			"For detailed documentation please visit https://istio-workspace-docs.netlify.com/\n\n",
		BashCompletionFunction: completion.BashCompletionFunc,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if released {
				go func() {
					latestRelease, _ := version.LatestRelease()
					if !version.IsLatestRelease(latestRelease) {
						releaseInfo <- fmt.Sprintf("WARN: you are using %s which is not the latest release (newest is %s).\n"+
							"Follow release notes for update info https://github.com/Maistra/istio-workspace/releases/latest", v.Version, latestRelease)
					} else {
						releaseInfo <- ""
					}
				}()
			}

			hook.Listen()
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				if flag.Changed && strings.Join(flag.Annotations["silent"], "") == "true" {
					log.SetLogger(log.CreateOperatorAwareLoggerWithLevel("root", zapcore.ErrorLevel))
				}
			})

			return errors.Wrap(config.SetupConfigSources(loadConfigFileName(cmd)), "failed setting config sources")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			shouldPrintVersion, _ := cmd.Flags().GetBool("version")
			if shouldPrintVersion {
				version.LogVersion()
			} else {
				fmt.Print(cmd.UsageString())
			}

			return nil
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			defer func() {
				close(releaseInfo)
			}()
			if released {
				timer := time.NewTimer(2 * time.Second)
				select {
				case release := <-releaseInfo:
					logger().Info(release)
				case <-timer.C:
					// do nothing, just timeout
				}
			}

			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			hook.Close()
		},
	}

	rootCmd.PersistentFlags().
		StringVarP(&configFile, "config", "c", ".ike.config.yaml",
			fmt.Sprintf("config file (supported formats: %s)", strings.Join(config.SupportedExtensions(), ", ")))
	rootCmd.Flags().Bool("version", false, "prints the version number of ike cli")
	rootCmd.PersistentFlags().String("help-format", "standard", "prints help in asciidoc table")
	if err := rootCmd.PersistentFlags().MarkHidden("help-format"); err != nil {
		logger().Error(err, "failed while trying to hide a flag")
	}

	config.SetupConfig()
	format.EnhanceHelper(rootCmd)
	format.RegisterTemplateFuncs()

	return rootCmd
}

func loadConfigFileName(cmd *cobra.Command) (configFileName string, defaultConfigSource bool) {
	configFlag := cmd.Flag("config")
	configFileName = viper.GetString("config")
	if configFileName == "" {
		if configFlag.Changed {
			configFileName = configFlag.Value.String()
		} else {
			configFileName = configFlag.DefValue
		}
	}
	defaultConfigSource = configFlag.DefValue == configFileName

	return
}
