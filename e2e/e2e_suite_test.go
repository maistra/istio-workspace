package e2e_test

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/pkg/naming"
	"github.com/maistra/istio-workspace/pkg/shell"
	. "github.com/maistra/istio-workspace/test"
	testshell "github.com/maistra/istio-workspace/test/shell"

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

	ClientVersion()

	tmpPath.SetPath(path.Dir(shell.CurrentDir())+"/dist", os.Getenv("PATH"))
	ensureRequiredBinaries()

	shouldManageCluster := manageCluster()

	if shouldManageCluster {
		tmpClusterDir = TmpDir(GinkgoT(), "/tmp/ike-e2e-tests/cluster-maistra-"+naming.RandName(16))
		executeWithTimer(func() {
			fmt.Printf("\nStarting up Openshift/Istio cluster in [%s]\n", tmpClusterDir)
			projectDir := os.Getenv("PROJECT_DIR")
			Expect(os.Setenv("IKE_CLUSTER_DIR", tmpClusterDir)).ToNot(HaveOccurred())
			<-testshell.ExecuteInDir(projectDir, "make", "start-cluster").Done()
		})
	}

	executeWithTimer(func() {
		LoginAsTestPowerUser()

		fmt.Printf("\nExposing Docker Registry\n")
		if ClientVersion() == 4 {
			<-testshell.Execute(`oc patch configs.imageregistry.operator.openshift.io/cluster --patch '{"spec":{"defaultRoute":true}}' --type=merge`).Done()
		} else {
			<-testshell.Execute("oc expose service docker-registry -n default").Done()
		}

		if shouldManageCluster {
			LoadIstio() // Not needed for running cluster
		}

		_ = CreateOperatorNamespace()
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
		if manageCluster() {
			executeWithTimer(func() {
				fmt.Println("\nStopping Openshift/Istio cluster")
				testshell.Execute("oc cluster down")
			})
		}

		fmt.Printf("Don't forget to wipe out %s where test cluster sits\n", tmpClusterDir)
		fmt.Println("For example by using such command: ")
		fmt.Printf("$ mount | grep openshift | cut -d' ' -f 3 | xargs -I {} sudo umount {} && sudo rm -rf %s", tmpClusterDir)
	})

func createProjectsForCompletionTests() {
	LoginAsTestPowerUser()
	testshell.ExecuteAll(
		NewProjectCmd("datawire-project"),
		deployHelloWorldCmd("my-datawire-deployment"),
		NewProjectCmd("datawire-other-project"),
		deployHelloWorldCmd("other-1-datawire-deployment"),
		deployHelloWorldCmd("other-2-datawire-deployment"),
	)
}

func deleteProjectsForCompletionTests() {
	LoginAsTestPowerUser()
	testshell.ExecuteAll(
		"oc delete project datawire-project",
		"oc delete project datawire-other-project",
	)
}

func deployHelloWorldCmd(name string) string {
	return "oc new-app --docker-image datawire/hello-world --name " + name + " --allow-missing-images"
}

func manageCluster() bool {
	if manage, found := os.LookupEnv("IKE_E2E_MANAGE_CLUSTER"); found {
		shouldManage, _ := strconv.ParseBool(manage)
		return shouldManage
	}

	return true
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
