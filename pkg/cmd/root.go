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
	"github.com/maistra/istio-workspace/pkg/k8s"
	"github.com/maistra/istio-workspace/pkg/log"
	v "github.com/maistra/istio-workspace/version"
)

const (
	AnnotationOperatorRequired = "operator-required"

	DocsLink = "https://istio-workspace-docs.netlify.com"
)

var logger = func() logr.Logger {
	return log.Log.WithValues("type", "root")
}

// NewCmd creates instance of root "ike" Cobra Command with flags and execution logic defined.
func NewCmd(verifier k8s.InstallationVerifier) *cobra.Command {
	var configFile string
	releaseInfo := make(chan string, 1)
	released := v.Released()

	rootCmd := &cobra.Command{
		SilenceErrors: true,
		Use:           "ike",
		Short: `ike lets you safely develop and test on production without a sweat!

For detailed documentation please visit ` + DocsLink,
		BashCompletionFunction: completion.BashCompletionFunc,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var collectedErrors error

			if err := config.SetupConfigSources(loadConfigFileName(cmd)); err != nil {
				collectedErrors = errors.Append(collectedErrors, errors.Wrapf(err, "failed setting config sources"))
			}

			if released {
				go checkIfLatestRelease(releaseInfo)
			}

			if cmd.Annotations[AnnotationOperatorRequired] == "true" {
				crdExists, err := verifier.CheckCRD()
				if !crdExists {
					return errors.Wrapf(err, "failed to locate istio-operator on your cluster, "+
						"please follow installation instructions %s/istio-workspace/%s/getting_started.html#_installing_cluster_component\n", DocsLink, v.CurrentVersion())
				}
			}

			hook.Listen()
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				if flag.Changed && strings.Join(flag.Annotations["silent"], "") == "true" {
					log.SetLogger(log.CreateOperatorAwareLoggerWithLevel("root", zapcore.ErrorLevel))
				}
			})

			return errors.Wrap(collectedErrors, "failed setting up command")
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
				if releaseInfo != nil {
					close(releaseInfo)
				}
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
			hook.Close() // in case of error during cmd run invoking func needs to call it (see main.go)
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

func checkIfLatestRelease(releaseInfo chan<- string) {
	latestRelease, _ := version.LatestRelease()
	if !version.IsLatestRelease(latestRelease) {
		releaseInfo <- fmt.Sprintf("WARN: you are using %s which is not the latest release (newest is %s).\n"+
			"Follow release notes for update info https://github.com/Maistra/istio-workspace/releases/latest", v.Version, latestRelease)
	} else {
		releaseInfo <- ""
	}
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
