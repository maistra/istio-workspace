package e2e_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/aslakknutsen/istio-workspace/cmd/ike/cmd"
	"github.com/aslakknutsen/istio-workspace/pkg/naming"
	. "github.com/aslakknutsen/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "End To End Test Suite")
}

var _, skipCluster = os.LookupEnv("SKIP_CLUSTER")
var runCluster = !skipCluster

var tmpClusterDir string

var _ = SynchronizedBeforeSuite(func() []byte {

	if runCluster {
		ensureRequiredBinaries()
		rand.Seed(time.Now().UTC().UnixNano())
		tmpClusterDir = TmpDir(GinkgoT(), "/tmp/ike-e2e-tests/cluster-maistra-"+naming.RandName(16))
		executeWithTimer(func() {
			fmt.Printf("\nStarting up Openshift/Istio cluster in [%s]\n", tmpClusterDir)
			<-cmd.Execute("istiooc", "cluster", "up",
				"--enable", "'registry,router,persistent-volumes,istio,centos-imagestreams'",
				"--base-dir", tmpClusterDir+"/maistra.local.cluster",
			).Done()

			fmt.Printf("\nExposing Docker Registry")
			<-cmd.Execute("oc", "create", "route", "--service=docker-registry").Done()

			// create a 'real user' we can use to push to the DockerRegsitry
			fmt.Printf("\nAdd admin user")
			<-cmd.Execute("oc", "create", "user", "admin").Done()
			<-cmd.Execute("oc", "adm", "policy", "add-cluster-role-to-user", "cluster-admin", "admin").Done()
		})
	}
	return nil
},
	func(data []byte) {})

var _ = SynchronizedAfterSuite(func() {},
	func() {
		if runCluster {
			executeWithTimer(func() {
				fmt.Println("\nStopping Openshift/Istio cluster")
				cmd.Execute("oc", "cluster", "down")
			})
		}
		fmt.Printf("Don't forget to wipe out %s where test cluster sits\n", tmpClusterDir)
		fmt.Println("For example by using such command: ")
		fmt.Printf("$ mount | grep openshift | cut -d' ' -f 3 | xargs -I {} sudo umount {} && sudo rm -rf %s", tmpClusterDir)
	})

func ensureRequiredBinaries() {
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
