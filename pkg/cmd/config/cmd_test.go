package config_test

import (
	"fmt"

	. "github.com/maistra/istio-workspace/pkg/cmd"
	"github.com/maistra/istio-workspace/pkg/cmd/config"
	. "github.com/maistra/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"
)

const (
	otherValue     = "VALUE_OTHER"
	expectedValue  = "VALUE_SET"
	expectedOutput = "SUCCESS"
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

		It("should load from command name env context", func() {
			defer TemporaryEnvVars("IKE_TEST_ARG", expectedValue)()
			output, err := Run(testCmd).Passing()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(And(ContainSubstring(expectedOutput), ContainSubstring(expectedValue)))
		})

		It("should load from global env context", func() {
			defer TemporaryEnvVars("IKE_ARG", expectedValue)()
			output, err := Run(testCmd).Passing()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(And(ContainSubstring(expectedOutput), ContainSubstring(expectedValue)))
		})

	})

	Context("load from config file", func() {

		It("should load from command name env context", func() {
			config := fmt.Sprintf(`test:
    arg: %v`, expectedValue)

			configFile := TmpFile(GinkgoT(), "env_config.yaml", config)
			output, err := Run(testCmd).Passing("--config", configFile.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(And(ContainSubstring(expectedOutput), ContainSubstring(expectedValue)))
		})

		It("should load from global env context", func() {
			config := fmt.Sprintf(`arg: %v`, expectedValue)

			configFile := TmpFile(GinkgoT(), "env_config.yaml", config)
			output, err := Run(testCmd).Passing("--config", configFile.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(And(ContainSubstring(expectedOutput), ContainSubstring(expectedValue)))
		})

	})

	Context("override order", func() {

		AfterEach(func() {
			CleanUpTmpFiles(GinkgoT())
		})

		It("should use command name context over env global", func() {
			defer TemporaryEnvVars("IKE_ARG", otherValue)()
			defer TemporaryEnvVars("IKE_TEST_ARG", expectedValue)()
			output, err := Run(testCmd).Passing()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(And(ContainSubstring(expectedOutput), ContainSubstring(expectedValue)))
		})

		It("should use command name context over config global", func() {
			config := fmt.Sprintf(`arg: %v
test:
    arg: %v`, otherValue, expectedValue)

			configFile := TmpFile(GinkgoT(), "env_config.yaml", config)
			output, err := Run(testCmd).Passing("--config", configFile.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(And(ContainSubstring(expectedOutput), ContainSubstring(expectedValue)))
		})

		It("should use arguments context over command name env context", func() {
			defer TemporaryEnvVars("IKE_TEST_ARG", otherValue)()
			output, err := Run(testCmd).Passing("--arg", expectedValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(And(ContainSubstring(expectedOutput), ContainSubstring(expectedValue)))
		})

		It("should use arguments context over global env context", func() {
			defer TemporaryEnvVars("IKE_ARG", otherValue)()
			output, err := Run(testCmd).Passing("--arg", expectedValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(And(ContainSubstring(expectedOutput), ContainSubstring(expectedValue)))
		})

		It("should use arguments context over command name config context", func() {
			config := fmt.Sprintf(`test:
    arg: %v`, otherValue)

			configFile := TmpFile(GinkgoT(), "env_config.yaml", config)
			output, err := Run(testCmd).Passing("--config", configFile.Name(), "--arg", expectedValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(And(ContainSubstring(expectedOutput), ContainSubstring(expectedValue)))
		})

		It("should use arguments context over global config context", func() {
			config := fmt.Sprintf(`test:
    arg: %v`, otherValue)
			configFile := TmpFile(GinkgoT(), "env_config.yaml", config)
			output, err := Run(testCmd).Passing("--config", configFile.Name(), "--arg", expectedValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(And(ContainSubstring(expectedOutput), ContainSubstring(expectedValue)))
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
			value, _ := cmd.Flags().GetString("arg")
			cmd.Println(expectedOutput, ":", value)
			return nil
		},
	}

	testCmd.Flags().StringP("arg", "a", "", "test argument")

	_ = testCmd.MarkFlagRequired("arg")

	return testCmd
}
