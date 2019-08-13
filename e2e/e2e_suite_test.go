package e2e_test

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
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

var skipClusterShutdown bool

var tmpClusterDir string

var tmpPath = NewTmpPath()

var _ = SynchronizedBeforeSuite(func() []byte {
	tmpPath.SetPath(path.Dir(shell.CurrentDir())+"/dist", os.Getenv("PATH"))
	ensureRequiredBinaries()

	if skip, found := os.LookupEnv("IKE_E2E_SKIP_CLUSTER_SHUTDOWN"); found {
		skipClusterShutdown, _ = strconv.ParseBool(skip)
	}

	if clusterNotRunning() {
		rand.Seed(time.Now().UTC().UnixNano())
		tmpClusterDir = TmpDir(GinkgoT(), "/tmp/ike-e2e-tests/cluster-maistra-"+naming.RandName(16))
		executeWithTimer(func() {
			fmt.Printf("\nStarting up Openshift/Istio cluster in [%s]\n", tmpClusterDir)
			projectDir := os.Getenv("PROJECT_DIR")
			Expect(os.Setenv("IKE_CLUSTER_DIR", tmpClusterDir)).ToNot(HaveOccurred())
			<-testshell.ExecuteInDir(projectDir, "make", "start-cluster").Done()
		})
	} else {
		if _, found := os.LookupEnv("IKE_E2E_SKIP_CLUSTER_SHUTDOWN"); !found {
			skipClusterShutdown = true
		}
	}
	executeWithTimer(func() {
		<-testshell.Execute("oc login -u system:admin").Done()

		fmt.Printf("\nExposing Docker Registry\n")
		<-testshell.Execute("oc expose service docker-registry -n default").Done()

		// create a 'real user' we can use to push to the DockerRegistry
		fmt.Printf("\nAdd admin user\n")
		testshell.ExecuteAll(
			"oc create user admin",
			"oc adm policy add-cluster-role-to-user cluster-admin admin",
		)

		LoadIstio()

		LoginAsAdminUser()
		_ = CreateOperatorNamespace()
		BuildOperator()

		createProjectsForCompletionTests()
	})
	return nil
},
	func(data []byte) {})

func createProjectsForCompletionTests() {
	LoginAsAdminUser()
	testshell.ExecuteAll(
		"oc project myproject",
		deployHelloWorldCmd("my-datawire-deployment"),
		"oc new-project otherproject",
		deployHelloWorldCmd("other-1-datawire-deployment"),
		deployHelloWorldCmd("other-2-datawire-deployment"),
		"oc project myproject",
	)
}

func deployHelloWorldCmd(name string) string {
	return "oc new-app --docker-image datawire/hello-world --name " + name + " --allow-missing-images"
}

var _ = SynchronizedAfterSuite(func() {},
	func() {
		tmpPath.Restore()
		if !skipClusterShutdown {
			executeWithTimer(func() {
				fmt.Println("\nStopping Openshift/Istio cluster")
				testshell.Execute("oc cluster down")
			})
		}

		fmt.Printf("Don't forget to wipe out %s where test cluster sits\n", tmpClusterDir)
		fmt.Println("For example by using such command: ")
		fmt.Printf("$ mount | grep openshift | cut -d' ' -f 3 | xargs -I {} sudo umount {} && sudo rm -rf %s", tmpClusterDir)
	})

func clusterNotRunning() bool {
	clusterStatus := testshell.Execute("oc cluster status")
	<-clusterStatus.Done()
	return strings.Contains(strings.Join(clusterStatus.Status().Stdout, " "), "not running") // Error for this command is logged to stdout
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
