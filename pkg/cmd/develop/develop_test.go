package develop_test

import (
	"os"
	"path"

	"github.com/maistra/istio-workspace/pkg/cmd/develop"

	"github.com/maistra/istio-workspace/test/shell"

	. "github.com/maistra/istio-workspace/pkg/cmd"
	. "github.com/maistra/istio-workspace/test"

	"github.com/spf13/afero"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"
)

var _ = Describe("Usage of ike develop command", func() {

	var developCmd *cobra.Command

	BeforeEach(func() {
		developCmd = develop.NewCmd()
		developCmd.SilenceUsage = true
		developCmd.SilenceErrors = true
		NewCmd().AddCommand(developCmd)
	})

	Context("checking telepresence binary existence", func() {

		tmpPath := NewTmpPath()
		BeforeEach(func() {
			tmpPath.SetPath(path.Dir(shell.MvnBin), path.Dir(shell.TpSleepBin))
		})
		AfterEach(tmpPath.Restore)

		It("should fail invoking develop cmd when telepresence binary is not on $PATH", func() {
			oldPath := os.Getenv("PATH")
			_ = os.Setenv("PATH", "")
			defer func() {
				_ = os.Setenv("PATH", oldPath)
			}()

			_, err := ValidateArgumentsOf(developCmd).Passing("-r", "./test.sh", "-d", "hello-world")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("unable to find telepresence on your $PATH"))
		})

	})

	Describe("input validation", func() {

		tmpPath := NewTmpPath()
		BeforeEach(func() {
			tmpPath.SetPath(path.Dir(shell.MvnBin), path.Dir(shell.TpSleepBin))
		})
		AfterEach(tmpPath.Restore)

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
				_, err := ValidateArgumentsOf(developCmd).Passing("--deployment", "rating-service", "--run", "java -jar rating.jar")

				Expect(err).NotTo(HaveOccurred())
				Expect(developCmd.Flag("port").Value.String()).To(Equal("8000"))
			})

			It("should have default method inject-tcp when flag not specified", func() {
				_, err := ValidateArgumentsOf(developCmd).Passing("--deployment", "rating-service", "--run", "java -jar rating.jar")

				Expect(err).NotTo(HaveOccurred())
				Expect(developCmd.Flag("method").Value.String()).To(Equal("inject-tcp"))
			})

			It("should be able to provide the traffic route parameter", func() {
				_, err := ValidateArgumentsOf(developCmd).Passing("--deployment", "rating-service", "--run", "java", "--route", "header:name=value")

				Expect(err).NotTo(HaveOccurred())
				Expect(developCmd.Flag("route").Value.String()).To(Equal("header:name=value"))
			})

		})

		Context("with config file", func() {

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
				_, err := ValidateArgumentsOf(developCmd).Passing("--config", configFile.Name(), "--run", "'./test.sh'")

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

	Describe("telepresence arguments delegation", func() {

		tmpPath := NewTmpPath()
		BeforeEach(func() {
			tmpPath.SetPath(path.Dir(shell.MvnBin), path.Dir(shell.TpSleepBin))
		})
		AfterEach(tmpPath.Restore)

		It("should pass all specified parameters", func() {
			output, err := Run(developCmd).Passing("--deployment", "rating-service",
				"--run", "java -jar rating.jar",
				"--port", "4321:5000",
				"--method", "vpn-tcp",
				"--offline")

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("--deployment rating-service"))
			Expect(output).To(ContainSubstring("--expose 4321:5000"))
			Expect(output).To(ContainSubstring("--method vpn-tcp"))
			Expect(output).To(ContainSubstring("--run java -jar rating.jar"))
		})

		It("should pass specified parameters and defaults", func() {
			output, err := Run(developCmd).Passing("--deployment", "rating-service",
				"--run", "java -jar rating.jar",
				"--offline")

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("--deployment rating-service"))
			Expect(output).To(ContainSubstring("--expose 8000"))
			Expect(output).To(ContainSubstring("--method inject-tcp"))
			Expect(output).To(ContainSubstring("--run java -jar rating.jar"))
		})

		It("should pass specified parameters and defaults", func() {
			output, err := Run(developCmd).Passing("--deployment", "rating-service",
				"--run", "java -jar rating.jar",
				"--namespace", "my-project",
				"--offline")

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("--deployment rating-service"))
			Expect(output).To(ContainSubstring("--namespace my-project"))
			Expect(output).To(ContainSubstring("--expose 8000"))
			Expect(output).To(ContainSubstring("--method inject-tcp"))
			Expect(output).To(ContainSubstring("--run java -jar rating.jar"))
		})

	})

	Context("build execution", func() {

		tmpPath := NewTmpPath()
		BeforeEach(func() {
			tmpPath.SetPath(path.Dir(shell.TpBin), path.Dir(shell.MvnBin))
		})
		AfterEach(tmpPath.Restore)

		It("should execute build when specified", func() {
			output, err := Run(developCmd).Passing("--deployment", "rating-service",
				"--run", "java -jar rating.jar",
				"--build", "mvn clean install",
				"--port", "4321",
				"--method", "vpn-tcp",
				"--offline")

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("mvn clean install"))

		})

		It("should not execute build when --no-build specified", func() {
			output, err := Run(developCmd).Passing("--deployment", "rating-service",
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
