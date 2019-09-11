package version

import (
	"fmt"
	"runtime"

	"github.com/maistra/istio-workspace/version"

	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/spf13/cobra"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("version")

// NewCmd creates version cmd which prints version and Build details of the executed binary
func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "Prints the version number of ike cli",
		Long:         "All software has versions. This is Ike's",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			PrintVersion()
			return nil
		},
	}
}

func PrintVersion() {
	log.Info(fmt.Sprintf("Ike Version: %s", version.Version))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("operator-sdk Version: %v", sdkVersion.Version))
	log.Info(fmt.Sprintf("Build Commit: %v", version.Commit))
	log.Info(fmt.Sprintf("Build Time: %v", version.BuildTime))
}
