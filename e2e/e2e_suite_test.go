package e2e_test

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/maistra/istio-workspace/cmd/ike/cmd"
	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/pkg/naming"
	. "github.com/maistra/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "End To End Test Suite")
}

var _, skipClusterShutdown = os.LookupEnv("SKIP_CLUSTER_SHUTDOWN")

var tmpClusterDir string

var tmpPath = NewTmpPath()

var _ = SynchronizedBeforeSuite(func() []byte {
	tmpPath.SetPath(path.Dir(cmd.CurrentDir())+"/dist", os.Getenv("PATH"))
	ensureRequiredBinaries()
	if clusterNotRunning() {
		rand.Seed(time.Now().UTC().UnixNano())
		tmpClusterDir = TmpDir(GinkgoT(), "/tmp/ike-e2e-tests/cluster-maistra-"+naming.RandName(16))
		executeWithTimer(func() {
			fmt.Printf("\nStarting up Openshift/Istio cluster in [%s]\n", tmpClusterDir)
			<-cmd.ExecuteInDir(".", "istiooc", "cluster", "up",
				"--enable", "registry,router,persistent-volumes,istio,centos-imagestreams",
				"--base-dir", tmpClusterDir+"/maistra.local.cluster",
			).Done()
		})
		skipClusterShutdown = true
	}
	executeWithTimer(func() {
		<-cmd.Execute("oc login -u system:admin").Done()

		fmt.Printf("\nExposing Docker Registry\n")
		<-cmd.Execute("oc expose service docker-registry -n default").Done()

		// create a 'real user' we can use to push to the DockerRegistry
		fmt.Printf("\nAdd admin user\n")
		<-cmd.Execute("oc create user admin").Done()
		<-cmd.Execute("oc adm policy add-cluster-role-to-user cluster-admin admin").Done()

		LoadIstio()

		<-cmd.Execute("oc login -u admin -p admin").Done()
		workspaceNamespace := CreateOperatorNamespace()
		BuildOperator()
		DeployOperator()
		Eventually(AllPodsNotInState(workspaceNamespace, "Running"), 3*time.Minute, 2*time.Second).
			Should(ContainSubstring("No resources found"))
	})
	return nil
},
	func(data []byte) {})

var _ = SynchronizedAfterSuite(func() {},
	func() {
		tmpPath.Restore()
		if !skipClusterShutdown {
			executeWithTimer(func() {
				fmt.Println("\nStopping Openshift/Istio cluster")
				cmd.Execute("oc cluster down")
			})
		}
		fmt.Printf("Don't forget to wipe out %s where test cluster sits\n", tmpClusterDir)
		fmt.Println("For example by using such command: ")
		fmt.Printf("$ mount | grep openshift | cut -d' ' -f 3 | xargs -I {} sudo umount {} && sudo rm -rf %s", tmpClusterDir)
	})

func clusterNotRunning() bool {
	clusterStatus := cmd.Execute("oc cluster status")
	<-clusterStatus.Done()
	return strings.Contains(strings.Join(clusterStatus.Status().Stdout, " "), "not running")
}

func ensureRequiredBinaries() {
	Expect(cmd.BinaryExists("ike", "make sure you have binary in the ./dist folder. Try make compile at least")).To(BeTrue())
	Expect(cmd.BinaryExists("istiooc", "check https://maistra.io/ for details")).To(BeTrue())
	Expect(cmd.BinaryExists("oc", "grab latest openshift origin client tools from here https://github.com/openshift/origin/releases")).To(BeTrue())
	Expect(cmd.BinaryExists("python3", "make sure you have python3 installed")).To(BeTrue())
}

type noArgFunc func()

func executeWithTimer(f noArgFunc) {
	start := time.Now()
	f()
	fmt.Printf("... done in %s\n", time.Since(start))
}
