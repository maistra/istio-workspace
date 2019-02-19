package cmd_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	. "github.com/aslakknutsen/istio-workspace/cmd/ike/cmd"
	. "github.com/aslakknutsen/istio-workspace/test"

	"github.com/onsi/gomega/gexec"
	"github.com/spf13/afero"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"
)

var _ = Describe("Usage of ike develop command", func() {

	var developCmd *cobra.Command

	var mvnBin string
	var tpBin string
	var tpSleepBin string

	BeforeEach(func() {
		developCmd = NewDevelopCmd()
		developCmd.SilenceUsage = true
		developCmd.SilenceErrors = true
		NewRootCmd().AddCommand(developCmd)
	})

	BeforeSuite(func() {
		mvnBin = buildBinary("github.com/aslakknutsen/istio-workspace/test/echo", "mvn")
		tpBin = buildBinary("github.com/aslakknutsen/istio-workspace/test/echo", "telepresence")
		tpSleepBin = buildBinary("github.com/aslakknutsen/istio-workspace/test/echo",
			"telepresence", "-ldflags", "-w -X main.SleepMs=50")

		fmt.Println(mvnBin)
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	Context("checking telepresence binary existence", func() {

		tmpPath := NewTmpPath()
		BeforeEach(func() {
			tmpPath.SetPath(path.Dir(mvnBin), path.Dir(tpSleepBin))
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
			tmpPath.SetPath(path.Dir(mvnBin), path.Dir(tpSleepBin))
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
			tmpPath.SetPath(path.Dir(mvnBin), path.Dir(tpSleepBin))
		})
		AfterEach(tmpPath.Restore)

		It("should pass all specified parameters", func() {
			output, err := Execute(developCmd).Passing("--deployment", "rating-service",
				"--run", "java -jar rating.jar",
				"--port", "4321",
				"--method", "vpn-tcp")

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("--swap-deployment rating-service"))
			Expect(output).To(ContainSubstring("--expose 4321"))
			Expect(output).To(ContainSubstring("--method vpn-tcp"))
			Expect(output).To(ContainSubstring("--run java -jar rating.jar"))
		})

		It("should pass specified parameters and defaults", func() {
			output, err := Execute(developCmd).Passing("--deployment", "rating-service",
				"--run", "java -jar rating.jar")

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("--swap-deployment rating-service"))
			Expect(output).To(ContainSubstring("--expose 8000"))
			Expect(output).To(ContainSubstring("--method inject-tcp"))
			Expect(output).To(ContainSubstring("--run java -jar rating.jar"))
		})

	})

	Context("build execution", func() {

		var originalPath string

		BeforeEach(func() {
			// we stub existence of telepresence executable as develop command does a precondition check before execution
			// to verify if it exists on the PATH
			originalPath = os.Getenv("PATH")
			_ = os.Setenv("PATH", path.Dir(tpBin)+":"+path.Dir(mvnBin))
		})

		AfterEach(func() {
			_ = os.Setenv("PATH", originalPath)
		})

		It("should execute build when specified", func() {
			output, err := Execute(developCmd).Passing("--deployment", "rating-service",
				"--run", "java -jar rating.jar",
				"--build", "mvn clean install",
				"--port", "4321",
				"--method", "vpn-tcp")
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(ContainSubstring("mvn clean install"))

		})

		It("should not execute build when --no-build specified", func() {
			output, err := Execute(developCmd).Passing("--deployment", "rating-service",
				"--run", "java -jar rating.jar",
				"--build", "mvn clean install",
				"--no-build",
				"--port", "4321",
				"--method", "vpn-tcp")

			Expect(err).NotTo(HaveOccurred())
			Expect(output).ToNot(ContainSubstring("mvn clean install"))
		})

	})

	Context("watching file changes", func() {

		tmpPath := NewTmpPath()
		BeforeEach(func() {
			tmpPath.SetPath(path.Dir(mvnBin), path.Dir(tpSleepBin))
		})
		AfterEach(tmpPath.Restore)

		It("should re-build and re-run telepresence", func() {
			// given
			config := TmpFile(GinkgoT(), "/tmp/watch-test/config.yaml", "content")
			outputChan := make(chan string)
			go func() {
				defer GinkgoRecover()
				output, err := Execute(developCmd).Passing("--deployment", "rating-service",
					"--run", "java -jar rating.jar",
					"--build", "mvn clean install",
					"--port", "4321",
					"--watch", "/tmp/watch-test",
					"--method", "vpn-tcp")
				Expect(err).NotTo(HaveOccurred())
				outputChan <- output
			}()

			// when
			time.Sleep(25 * time.Millisecond) // as tp process sleeps for 50ms, we wait before we start modifying the file

			_, _ = config.WriteString("modified!")

			// then
			var output string
			Eventually(outputChan).Should(Receive(&output))
			Expect(output).To(ContainSubstring("config.yaml changed. Restarting process."))
			Expect(strings.Count(output, "mvn clean install")).To(Equal(2))
		})

		It("should ignore build if not defined and just re-run telepresence", func() {
			config := TmpFile(GinkgoT(), "/tmp/watch-test/config.yaml", "content")

			outputChan := make(chan string)
			go func() {
				defer GinkgoRecover()
				output, err := Execute(developCmd).Passing("--deployment", "rating-service",
					"--run", "java -jar rating.jar",
					"--port", "4321",
					"--watch", "/tmp/watch-test",
					"--method", "vpn-tcp")
				Expect(err).NotTo(HaveOccurred())
				outputChan <- output
			}()

			time.Sleep(25 * time.Millisecond) // as tp process sleeps for 50ms, we wait before we start modifying the file
			_, _ = config.WriteString("modified!")

			var output string
			Eventually(outputChan).Should(Receive(&output))
			Expect(output).To(ContainSubstring("config.yaml changed. Restarting process."))
			Expect(strings.Count(output, "mvn clean install")).To(Equal(0))
		})

		It("should only re-run telepresence when --no-build flag specified", func() {
			config := TmpFile(GinkgoT(), "/tmp/watch-test/config.yaml", "content")

			outputChan := make(chan string)
			go func() {
				defer GinkgoRecover()
				output, err := Execute(developCmd).Passing("--deployment", "rating-service",
					"--run", "java -jar rating.jar",
					"--build", "mvn clean install",
					"--no-build", // TODO source from config
					"--port", "4321",
					"--watch", "/tmp/watch-test",
					"--method", "vpn-tcp")
				Expect(err).NotTo(HaveOccurred())
				outputChan <- output
			}()

			time.Sleep(25 * time.Millisecond) // as tp process sleeps for 50ms, we wait before we start modifying the file
			_, _ = config.WriteString("modified!")

			var output string
			Eventually(outputChan).Should(Receive(&output))
			Expect(output).To(ContainSubstring("config.yaml changed. Restarting process."))
			Expect(strings.Count(output, "mvn clean install")).To(Equal(0))
		})
	})

})

var appFs = afero.NewOsFs()

func buildBinary(packagePath, name string, flags ...string) string {

	binPath, err := gexec.Build(packagePath, flags...)
	Expect(err).ToNot(HaveOccurred())

	// gexec.Build from Ginkgo does not allow to specify `-o` flag for the final binary name
	// thus we rename the binary instead. TODO: pull request to ginkgo
	if name != path.Base(packagePath) {
		finalName := copyBinary(appFs, binPath, name)
		_ = os.Remove(binPath)
		return finalName
	}

	return binPath
}

func copyBinary(appFs afero.Fs, src, dest string) string {
	binPath := path.Dir(src) + "/" + dest
	bin, err := appFs.Create(binPath)
	Expect(err).ToNot(HaveOccurred())

	err = appFs.Chmod(binPath, os.ModePerm)
	Expect(err).ToNot(HaveOccurred())

	content, err := afero.ReadFile(appFs, src)
	Expect(err).ToNot(HaveOccurred())
	_, err = bin.Write(content)
	Expect(err).ToNot(HaveOccurred())

	defer func() {
		err = bin.Close()
		Expect(err).ToNot(HaveOccurred())
	}()

	return bin.Name()
}
