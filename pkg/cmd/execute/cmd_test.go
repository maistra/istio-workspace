package execute_test

import (
	"path"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	. "github.com/maistra/istio-workspace/pkg/cmd"
	"github.com/maistra/istio-workspace/pkg/cmd/execute"
	. "github.com/maistra/istio-workspace/test"
	"github.com/maistra/istio-workspace/test/shell"
)

var _ = Describe("Usage of ike execute command", func() {

	var executeCmd *cobra.Command
	tmpFs := NewTmpFileSystem(GinkgoT())

	BeforeEach(func() {
		executeCmd = execute.NewCmd()
		executeCmd.SilenceUsage = true
		executeCmd.SilenceErrors = false
		executeCmd.Annotations = make(map[string]string, 1)
		executeCmd.Annotations["test"] = "true"
		NewCmd().AddCommand(executeCmd)
	})

	AfterEach(func() {
		tmpFs.Cleanup()
	})

	Context("watching file changes", func() {

		tmpPath := NewTmpPath()

		BeforeEach(func() {
			tmpPath.SetPath(path.Dir(shell.MvnBin), path.Dir(shell.Tp1WithSleepBin), path.Dir(shell.JavaBin))
		})

		AfterEach(tmpPath.Restore)

		It("should re-build and re-run java process", func() {
			// given
			tmpDir := tmpFs.Dir("re-run")
			code := tmpFs.File(tmpDir+"/watch-test/rating.java", "content")
			outputChan := make(chan string)

			go shell.ExecuteCommand(outputChan, func() (string, error) {
				return Run(executeCmd).Passing(
					"--run", "java -jar rating.jar",
					"--build", "mvn clean install",
					"--watch",
					"--dir", tmpDir+"/watch-test",
					// for testing purposes we handle file change events more frequently to avoid excessively long tests
					"--interval", "10",
				)
			})()
			// when
			time.Sleep(25 * time.Millisecond)
			_, _ = code.WriteString("modified!")
			time.Sleep(50 * time.Millisecond)
			simulateSigterm(executeCmd)

			// then
			var output string
			Eventually(outputChan).Should(Receive(&output))
			Expect(output).To(ContainSubstring("rating.java changed. Restarting process."))
			Expect(strings.Count(output, "mvn clean install")).To(Equal(2), "Expected build to be re-run.")
			Expect(strings.Count(output, "java -jar rating.jar")).To(Equal(2), "Expected process to be restarted.")
		})

		It("should start java process once if only log file is changing", func() {
			// given
			tmpDir := tmpFs.Dir("start-java")
			logFile := tmpFs.File(tmpDir+"/watch-test/tomcat.log", "content")
			outputChan := make(chan string)

			go shell.ExecuteCommand(outputChan, func() (string, error) {
				return Run(executeCmd).Passing(
					"--run", "java -jar rating.jar",
					"--watch",
					"--dir", tmpDir+"/watch-test",
					// for testing purposes we handle file change events more frequently to avoid excessively long tests
					"interval", "10",
				)
			})()

			// when
			time.Sleep(25 * time.Millisecond)
			_, _ = logFile.WriteString("\n>>> Server started")
			time.Sleep(50 * time.Millisecond)
			simulateSigterm(executeCmd)

			// then
			var output string
			Eventually(outputChan).Should(Receive(&output))
			Expect(output).ToNot(ContainSubstring("rating.java changed. Restarting process."))
			Expect(strings.Count(output, "java -jar rating.jar")).To(Equal(1), "Expected process to be executed once.")
		})

		It("should build and run java process only initially when changing file is excluded", func() {
			// given
			tmpDir := tmpFs.Dir("build-run-java-excluded")
			code := tmpFs.File(tmpDir+"/watch-test/rating.java", "content")
			outputChan := make(chan string)
			go shell.ExecuteCommand(outputChan, func() (string, error) {
				return Run(executeCmd).Passing(
					"--run", "java -jar rating.jar",
					"--watch",
					"--build", "mvn clean install",
					"--dir", tmpDir+"/watch-test",
					"--exclude", "*.java",
					// for testing purposes we handle file change events more frequently to avoid excessively long tests
					"--interval", "10",
				)
			})()

			// when
			time.Sleep(25 * time.Millisecond) // as tp process sleeps for 50ms, we wait before we start modifying the file
			_, _ = code.WriteString("modified!")
			time.Sleep(50 * time.Millisecond)
			simulateSigterm(executeCmd)

			// then
			var output string
			Eventually(outputChan).Should(Receive(&output))
			Expect(output).ToNot(ContainSubstring("rating.java changed. Restarting process."))
			Expect(strings.Count(output, "mvn clean install")).To(Equal(1), "Expected process to be started once.")
			Expect(strings.Count(output, "java -jar rating.jar")).To(Equal(1), "Expected build to be executed once.")
		})

		It("should ignore build if not defined and just re-run java process on file change", func() {
			// given
			tmpDir := tmpFs.Dir("ignore-build")
			code := tmpFs.File(tmpDir+"/watch-test/rating.java", "content")

			outputChan := make(chan string)
			go shell.ExecuteCommand(outputChan, func() (string, error) {
				return Run(executeCmd).Passing(
					"--run", "java -jar rating.jar",
					"--watch",
					"--dir", tmpDir+"/watch-test",
					// for testing purposes we handle file change events more frequently to avoid excessively long tests
					"--interval", "10",
				)
			})()

			// when
			time.Sleep(25 * time.Millisecond) // as tp process sleeps for 50ms, we wait before we start modifying the file
			_, _ = code.WriteString("modified!")
			time.Sleep(50 * time.Millisecond)
			simulateSigterm(executeCmd)

			// then
			var output string
			Eventually(outputChan).Should(Receive(&output))
			Expect(output).To(ContainSubstring("rating.java changed. Restarting process."))
			Expect(strings.Count(output, "mvn clean install")).To(Equal(0), "Expected build to not be executed.")
			Expect(strings.Count(output, "java -jar rating.jar")).To(Equal(2), "Expected process to be restarted.")
		})

		It("should only re-run java process when --no-build flag specified but build defined in config", func() {
			// given
			configFile := tmpFs.File("config.yaml", `watch:
  run: "java -jar config.jar"
  build: "mvn clean install"
`)
			tmpDir := tmpFs.Dir("re-run-no-build")
			code := tmpFs.File(tmpDir+"/watch-test/rating.java", "content")

			outputChan := make(chan string)
			go shell.ExecuteCommand(outputChan, func() (string, error) {
				return Run(executeCmd).Passing(
					"--config", configFile.Name(),
					"--watch",
					"--no-build",
					"--dir", tmpDir+"/watch-test",
					// for testing purposes we handle file change events more frequently to avoid excessively long tests
					"--interval", "10",
				)
			})()

			// when
			time.Sleep(25 * time.Millisecond) // as tp process sleeps for 50ms, we wait before we start modifying the file
			_, _ = code.WriteString("modified!")
			time.Sleep(50 * time.Millisecond)
			simulateSigterm(executeCmd)

			// then
			var output string
			Eventually(outputChan).Should(Receive(&output))
			Expect(output).To(ContainSubstring("rating.java changed. Restarting process."))
			Expect(strings.Count(output, "mvn clean install")).To(Equal(0), "Expected build to not be executed.")
		})
	})

})

func simulateSigterm(cmd *cobra.Command) {
	time.Sleep(16 * time.Millisecond)
	_ = cmd.Flag("kill").Value.Set("true")
}
