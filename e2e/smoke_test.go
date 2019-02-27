package e2e_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/aslakknutsen/istio-workspace/cmd/ike/cmd"
	"github.com/aslakknutsen/istio-workspace/e2e"
	"github.com/aslakknutsen/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Smoke End To End Tests", func() {

	Context("using ike develop against OpenShift Cluster with Istio (maistra)", func() {

		tmpPath := test.NewTmpPath()

		var (
			appName,
			tmpDir string
		)

		BeforeEach(func() {
			tmpPath.SetPath(os.Getenv("PATH"), path.Dir(cmd.CurrentDir())+"/dist")
			appName = randName(16)
			tmpDir = test.TmpDir(GinkgoT(), "app-"+appName)
		})

		AfterEach(func() {
			tmpPath.Restore()
			<-execute("oc", "delete", "all", "-l", "app="+appName).Done()
		})

		It("should replace python service with modified response", func() {

			createNewApp(appName)
			Eventually(callGetOn(appName), 1*time.Minute).Should(Equal("Hello, world!\n"))

			modifiedServerCode(tmpDir)
			ike := executeInDir(tmpDir, "ike", "develop",
				"--deployment", appName,
				"--port", "8000",
				"--method", "inject-tcp",
				"--run", "python3 server.py",
			)

			Eventually(callGetOn(appName), 3*time.Minute, 200*time.Millisecond).Should(Equal("Hello, telepresence! Ike Here!\n"))
			Expect(ike.Stop()).ToNot(HaveOccurred())
			Eventually(ike.Done(), 1*time.Minute).Should(BeClosed())
		})

		It("should watch python code changes and replace service when they occur", func() {

			createNewApp(appName)
			Eventually(callGetOn(appName), 1*time.Minute).Should(Equal("Hello, world!\n"))

			originalServerCode(tmpDir)
			ikeWithWatch := executeInDir(tmpDir, "ike", "develop",
				"--deployment", appName,
				"--port", "8000",
				"--method", "inject-tcp",
				"--watch",
				"--run", "python3 server.py",
			)
			Eventually(callGetOn(appName), 3*time.Minute, 200*time.Millisecond).Should(Equal("Hello, world!\n"))

			modifiedServerCode(tmpDir)

			Eventually(callGetOn(appName), 3*time.Minute, 200*time.Millisecond).Should(Equal("Hello, telepresence! Ike Here!\n"))

			Expect(ikeWithWatch.Stop()).ToNot(HaveOccurred())
			Eventually(ikeWithWatch.Done(), 1*time.Minute).Should(BeClosed())
		})

	})
})

var appFs = afero.NewOsFs()

func createNewApp(name string) {
	<-execute("oc", "new-project", name).Done()
	<-execute("oc", "login", "-u", "system:admin").Done()
	<-execute("oc", "adm", "policy", "add-scc-to-user", "anyuid", "-z", "default", "-n", name).Done()
	<-execute("oc", "adm", "policy", "add-scc-to-user", "privileged", "-z", "default", "-n", name).Done()
	<-execute("oc", "login", "-u", "developer").Done()
	<-execute("oc", "new-app",
		"--docker-image", "datawire/hello-world",
		"--name", name,
		"--allow-missing-images",
	).Done()
	<-execute("oc", "expose", "svc/"+name).Done()
}

func modifiedServerCode(tmpDir string) {
	createFile(tmpDir+"/"+"server.py", e2e.ModifiedServerPy)
}

func originalServerCode(tmpDir string) {
	createFile(tmpDir+"/"+"server.py", e2e.OrigServerPy)
}

func createFile(filePath, content string) {
	file, err := appFs.Create(filePath)
	Expect(err).NotTo(HaveOccurred())
	err = appFs.Chmod(filePath, os.ModePerm)
	Expect(err).ToNot(HaveOccurred())
	_, err = file.WriteString(content)
	Expect(err).ToNot(HaveOccurred())
	defer func() {
		err = file.Close()
		Expect(err).ToNot(HaveOccurred())
	}()
}

func callGetOn(name string) func() string {
	return func() string {
		resp, err := http.Get(fmt.Sprintf("http://%[1]s-%[1]s.127.0.0.1.nip.io", name))
		if err != nil {
			return ""
		}
		defer resp.Body.Close()
		content, _ := ioutil.ReadAll(resp.Body)
		return string(content)
	}
}

// TODO make it shared as soon as https://github.com/aslakknutsen/istio-workspace/pull/32 gets merged
var letters = []rune("abcdefghijklmnopqrstuvwxyz")
var alphaNumeric = []rune("abcdefghijklmnopqrstuvwxyz0987654321-")

//  Must be an a lower case alphanumeric (a-z, and 0-9) string with a maximum length of 58 characters,
//  where the first character is a letter (a-z), and the '-'
//  character is allowed anywhere except the first or last character.
func randName(length int) string {
	if length > 58 {
		length = 58
	}

	b := make([]rune, length)
	for i := range b {
		b[i] = alphaNumeric[rand.Intn(len(alphaNumeric))]
	}
	b[0] = letters[rand.Intn(len(letters))]
	b[length-1] = letters[rand.Intn(len(letters))]
	return string(b)
}
