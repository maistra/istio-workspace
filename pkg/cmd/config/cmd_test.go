package config_test

import (
	. "github.com/maistra/istio-workspace/pkg/cmd"
	"github.com/maistra/istio-workspace/pkg/cmd/config"
	. "github.com/maistra/istio-workspace/test"

	"github.com/spf13/afero"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"
)

var _ = Describe("Usage of ike command configuration", func() {

	var testCmd *cobra.Command

	BeforeEach(func() {
		testCmd = NewTestCmd()
		testCmd.SilenceUsage = true
		testCmd.SilenceErrors = true
		NewCmd().AddCommand(testCmd)
	})

	Context("load from environment", func() {

		XIt("should load from command name env context", func() {})
		XIt("should load from global env context", func() {})
		XIt("should override command name context over global", func() {})
	})
	Context("load from config file", func() {

		XIt("should load from command name env context", func() {})
		XIt("should load from global env context", func() {})
		XIt("should override command name context over global", func() {})

	})
	Context("load from arguments", func() {

		XIt("should override arguments context over global", func() {})

	})

	Context("override order", func() {})

	Context("checking telepresence binary existence", func() {

		It("should fail invoking develop cmd when telepresence binary is not on $PATH", func() {
			_, err := ValidateArgumentsOf(testCmd).Passing("-r", "./test.sh", "-d", "hello-world")

		})

		const config = `develop:
  deployment: test
  run: "java -jar config.jar"
  port: 9876
`
		var configFile afero.File

		BeforeEach(func() {
			configFile = TmpFile(GinkgoT(), "config.yaml", config)
		})

		It("should fail when passing non-existing config file", func() {
			_, err := ValidateArgumentsOf(testCmd).Passing("--config", "~/test.yaml")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`Config File "test" Not Found`))
		})

		It("should not execute build when --no-build specified", func() {
			output, err := Run(testCmd).Passing("--deployment", "rating-service",
				"--run", "java -jar rating.jar",
				"--build", "mvn clean install",
				"--no-build",
				"--port", "4321",
				"--method", "vpn-tcp",
				"--offline")

			Expect(err).NotTo(HaveOccurred())
			Expect(output).ToNot(ContainSubstring("mvn clean install"))
		})

	})

})

func NewTestCmd() *cobra.Command {
	testCmd := &cobra.Command{
		Use:          "test",
		Short:        "Test of configuration",
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			return config.SyncFullyQualifiedFlags(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			return nil
		},
	}

	testCmd.Flags().StringP("deployment", "d", "", "name of the deployment or deployment config")
	testCmd.Flags().StringP("port", "p", "8000", "port to be exposed in format local[:remote]")

	_ = testCmd.MarkFlagRequired("deployment")

	return testCmd
}
