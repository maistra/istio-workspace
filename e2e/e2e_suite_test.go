package e2e_test

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/pkg/naming"
	"github.com/maistra/istio-workspace/pkg/shell"
	. "github.com/maistra/istio-workspace/test"
	testshell "github.com/maistra/istio-workspace/test/shell"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "End To End Test Suite")
}

const (
	PreparedImageV1 = "prepared-image"
	PreparedImageV2 = "image-prepared"
)

var tmpClusterDir string

var tmpPath = NewTmpPath()

var _ = SynchronizedBeforeSuite(func() []byte {
	rand.Seed(time.Now().UTC().UnixNano())

	if envFile, found := os.LookupEnv("ENV_FILE"); found {
		if err := godotenv.Overload(filepath.Join(testshell.GetProjectDir(), envFile)); err != nil {
			Fail(err.Error())
		}
	}

	tmpPath.SetPath(path.Dir(shell.CurrentDir())+"/dist", os.Getenv("PATH"))
	ensureRequiredBinaries()

	executeWithTimer(func() {
		if RunsOnOpenshift {
			LoginAsTestPowerUser()

			// Don't setup shared image namespace if we're not building the images as part of the test flow
			if !UsePrebuiltImages() {
				fmt.Printf("\nExposing Container Registry\n")

				<-testshell.Execute(`oc patch configs.imageregistry.operator.openshift.io/cluster --patch '{"spec":{"defaultRoute":true}}' --type=merge`).Done()

				<-testshell.Execute(NewProjectCmd(GetRepositoryName())).Done()
				UpdateSecurityConstraintsFor(GetRepositoryName())
			}
		}

		// Assume images are built by some external means
		if !UsePrebuiltImages() {
			BuildOperator()
			BuildTestService()
			BuildTestServicePreparedImage(PreparedImageV1)
			BuildTestServicePreparedImage(PreparedImageV2)
		}
		createProjectsForCompletionTests()
	})

	return nil
},
	func(data []byte) {})

var _ = SynchronizedAfterSuite(func() {},
	func() {
		deleteProjectsForCompletionTests()
		tmpPath.Restore()

		fmt.Printf("Don't forget to wipe out %s cluster directory!\n", tmpClusterDir)
		fmt.Println("For example by using such command: ")
		fmt.Printf("$ mount | grep openshift | cut -d' ' -f 3 | xargs -I {} sudo umount {} && sudo rm -rf %s", tmpClusterDir)
	})

var CompletionProject1 = "ike-autocompletion-test-" + naming.RandName(16)
var CompletionProject2 = "ike-autocompletion-test-" + naming.RandName(16)

func createProjectsForCompletionTests() {
	testshell.ExecuteAll(
		NewProjectCmd(CompletionProject1),
		NewProjectCmd(CompletionProject2),
	)
	testshell.ExecuteAll(DeployNoopLoopCmd("my-deployment", CompletionProject1)...)
	testshell.ExecuteAll(DeployNoopLoopCmd("other-1-deployment", CompletionProject2)...)
	testshell.ExecuteAll(DeployNoopLoopCmd("other-2-deployment", CompletionProject2)...)
}

func deleteProjectsForCompletionTests() {
	testshell.ExecuteAll(
		DeleteProjectCmd(CompletionProject1),
		DeleteProjectCmd(CompletionProject2),
	)
}

func ensureRequiredBinaries() {
	Expect(shell.BinaryExists("ike", "make sure you have binary in the ./dist folder. Try make compile at least")).To(BeTrue())
	ocExists := shell.BinaryExists("oc", "")
	kubectlExists := shell.BinaryExists("kubectl", "")
	Expect(kubectlExists || ocExists).To(BeTrue(), "make sure you have oc or kubectl installed")
	Expect(shell.BinaryExists("python3", "make sure you have python3 installed")).To(BeTrue())
}

type noArgFunc func()

func executeWithTimer(f noArgFunc) {
	start := time.Now()
	f()
	fmt.Printf("... done in %s\n", time.Since(start))
}
