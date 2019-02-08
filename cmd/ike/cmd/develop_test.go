package cmd_test

import (
	"os"

	. "github.com/aslakknutsen/istio-workspace/cmd/ike/cmd"

	. "github.com/aslakknutsen/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var _ = Describe("Usage of ike develop command", func() {

	var developCmd *cobra.Command

	BeforeEach(func() {
		developCmd = NewDevelopCmd()
		developCmd.SilenceUsage = false
		developCmd.SilenceErrors = false
		NewRootCmd().AddCommand(developCmd)
	})

	Context("with flags only", func() {

		It("should fail when deployment is not specified", func() {
			_, err := ValidateArgumentsOf(developCmd).Passing("--port", "1234")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(And(ContainSubstring("required flag(s)"), ContainSubstring("deployment")))
		})

		It("should fail when run command is not specified", func() {
			_, err := ValidateArgumentsOf(developCmd).Passing("--deployment", "rating-service")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(And(ContainSubstring("required flag(s)"), ContainSubstring("run")))
		})

		It("should have default port 8000 when flag not specified", func() {
			_, err := ValidateArgumentsOf(developCmd).Passing("--deployment", "rating-service", "--run", "'python3 rating.py'")

			Expect(err).NotTo(HaveOccurred())
			Expect(developCmd.Flag("port").Value.String()).To(Equal("8000"))
		})

		It("should have default method inject-tcp when flag not specified", func() {
			_, err := ValidateArgumentsOf(developCmd).Passing("--deployment", "rating-service", "--run", "'python3 rating.py'")

			Expect(err).NotTo(HaveOccurred())
			Expect(developCmd.Flag("method").Value.String()).To(Equal("inject-tcp"))
		})

	})

	XContext("with config file", func() {

		const config = `develop:
  deployment: test
  run: "python3 server.py"
  port: 5555
`
		var configFile afero.File

		BeforeEach(func() {
			configFile = TmpFile(GinkgoT(), "config.yaml", config)
		})

		AfterEach(func() {
			CleanUp(GinkgoT())
		})

		It("should fail when passing non-existing config file", func() {
			_, err := ValidateArgumentsOf(developCmd).Passing("--config", "~/test.yaml")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`Config File "test" Not Found`))
		})

		It("should load deployment from config file if not passed as flag", func() {
			_, err := ValidateArgumentsOf(developCmd).Passing("--port", "1234", "--config", configFile.Name())

			Expect(err).ToNot(HaveOccurred())
			Expect(developCmd.Flag("deployment").Value.String()).To(Equal("test"))
		})

		It("should use run defined in the flag not from config file", func() {
			_, err := ValidateArgumentsOf(developCmd).Passing("-r", "'./test.sh'", "--config", configFile.Name())

			Expect(err).ToNot(HaveOccurred())
			Expect(developCmd.Flag("run").Value.String()).To(Equal(`'./test.sh'`))
		})

		Context("with ENV port variable", func() {

			var oldEnv string

			BeforeEach(func() {
				oldEnv = os.Getenv("IKE_DEVELOP_PORT")
				_ = os.Setenv("IKE_DEVELOP_PORT", "4321")
			})

			AfterEach(func() {
				_ = os.Setenv("IKE_DEVELOP_PORT", oldEnv)
			})

			It("should use environment variable over config file", func() {
				_, err := ValidateArgumentsOf(developCmd).Passing("--config", configFile.Name())

				Expect(err).ToNot(HaveOccurred())
				Expect(developCmd.Flag("port").Value.String()).To(Equal("4321"))
			})

			It("should use flag over environment variable", func() {
				_, err := ValidateArgumentsOf(developCmd).Passing("--port", "1111", "--config", configFile.Name())

				Expect(err).ToNot(HaveOccurred())
				Expect(developCmd.Flag("port").Value.String()).To(Equal("1111"))
			})

		})

	})

})
