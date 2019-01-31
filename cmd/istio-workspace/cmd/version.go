package cmd

import (
	"runtime"

	"github.com/aslakknutsen/istio-workspace/version"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version number of ike cli",
	Long:  `All software has versions. This is Ike's`,
	Run: func(cmd *cobra.Command, args []string) {
		printVersion()
	},
}

func printVersion() {
	logrus.Infof("Ike Version: %s", version.Version)
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
	logrus.Infof("Build Commit: %v", version.Commit)
	logrus.Infof("Build Time: %v", version.BuildTime)
}
