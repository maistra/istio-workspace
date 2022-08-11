package execute_test

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shirou/gopsutil/v3/process"

	"github.com/maistra/istio-workspace/test"
	testshell "github.com/maistra/istio-workspace/test/shell"
)

var _ = Describe("ike execute - managing spawned processes", func() {

	Context("using actual binary", func() {
		tmpFs := test.NewTmpFileSystem(GinkgoT())
		tmpDir := tmpFs.Dir("ike-execute")

		tmpPath := test.NewTmpPath()

		BeforeEach(func() {
			tmpPath.SetPath(path.Dir(sleepBin), testshell.GetProjectDir()+string(os.PathSeparator)+"dist")
		})

		AfterEach(tmpPath.Restore)

		It("should kill child process", func() {
			// given
			ikeExecute := testshell.ExecuteInDir(tmpDir, "ike", "execute",
				"--run", "sleepy",
			)

			time.AfterFunc(50*time.Millisecond, func() {
				_ = syscall.Kill(ikeExecute.Status().PID, syscall.SIGINT)
			})

			Eventually(ikeExecute.Done(), 100*time.Millisecond).Should(BeClosed())

			pid, exists, err := findPID(ikeExecute.Status().Stdout)
			Expect(err).To(Not(HaveOccurred()))
			Expect(exists).To(BeFalse(), fmt.Sprintf("child process [%d] should not be running", pid))
		})

	})

})

func findPID(stdout []string) (pid int64, exists bool, err error) {
	re := regexp.MustCompile(`{(.*?)}`)
	match := re.FindStringSubmatch(strings.Join(stdout, ""))
	pid, _ = strconv.ParseInt(match[1], 10, 32)
	exists, err = process.PidExists(int32(pid))

	return pid, exists, err
}
