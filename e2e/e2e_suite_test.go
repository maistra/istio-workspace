package e2e_test

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"

	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/pkg/naming"
	"github.com/maistra/istio-workspace/pkg/shell"
	. "github.com/maistra/istio-workspace/test"
	testshell "github.com/maistra/istio-workspace/test/shell"

	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "End To End Test Suite")
}

var tmpClusterDir string

var tmpPath = NewTmpPath()

var _ = SynchronizedBeforeSuite(func() []byte {
	rand.Seed(time.Now().UTC().UnixNano())

	if envFile, found := os.LookupEnv("ENV_FILE"); found {
		if err := godotenv.Load(testshell.GetProjectDir() + string(os.PathSeparator) + envFile); err != nil {
			Fail(err.Error())
		}
	}

	tmpPath.SetPath(path.Dir(shell.CurrentDir())+"/dist", os.Getenv("PATH"))
	ensureRequiredBinaries()

	executeWithTimer(func() {
		LoginAsTestPowerUser()

		fmt.Printf("\nExposing Docker Registry\n")
		<-testshell.Execute(`oc patch configs.imageregistry.operator.openshift.io/cluster --patch '{"spec":{"defaultRoute":true}}' --type=merge`).Done()

		BuildOperator()
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
	LoginAsTestPowerUser()
	testshell.ExecuteAll(
		NewProjectCmd(CompletionProject1),
		DeployHelloWorldCmd("my-datawire-deployment", CompletionProject1),
		NewProjectCmd(CompletionProject2),
		DeployHelloWorldCmd("other-1-datawire-deployment", CompletionProject2),
		DeployHelloWorldCmd("other-2-datawire-deployment", CompletionProject2),
	)
}

func deleteProjectsForCompletionTests() {
	LoginAsTestPowerUser()
	testshell.ExecuteAll(
		DeleteProjectCmd(CompletionProject1),
		DeleteProjectCmd(CompletionProject2),
	)
}

func ensureRequiredBinaries() {
	Expect(shell.BinaryExists("ike", "make sure you have binary in the ./dist folder. Try make compile at least")).To(BeTrue())
	Expect(shell.BinaryExists("oc", "grab latest openshift origin client tools from here https://github.com/openshift/origin/releases")).To(BeTrue())
	Expect(shell.BinaryExists("python3", "make sure you have python3 installed")).To(BeTrue())
}

type noArgFunc func()

func executeWithTimer(f noArgFunc) {
	start := time.Now()
	f()
	fmt.Printf("... done in %s\n", time.Since(start))
}
