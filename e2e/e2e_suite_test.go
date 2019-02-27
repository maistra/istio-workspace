package e2e_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aslakknutsen/istio-workspace/cmd/ike/cmd"
	. "github.com/aslakknutsen/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gocmd "github.com/go-cmd/cmd"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "End To End Test Suite")
}

var tmpClusterDir = TmpDir(GinkgoT(), "/tmp/ike-e2e-tests/cluster-maistra-" + randName(16))

var _ = SynchronizedBeforeSuite(func() []byte {

	ensureRequiredBinaries()

	executeWithTimer(func() {
		fmt.Printf("\nStarting up Openshift/Istio cluster in [%s]\n", tmpClusterDir)
		<-execute("istiooc", "cluster", "up",
			"--enable", "'registry,router,persistent-volumes,istio,centos-imagestreams'",
			"--base-dir", tmpClusterDir,
		).Done()
	})

	return nil
},
	func(data []byte) {})

var _ = SynchronizedAfterSuite(func() {},
	func() {
		executeWithTimer(func() {
			fmt.Println("\nStopping Openshift/Istio cluster")
			<-execute("oc", "cluster", "down").Done()
		})

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

func execute(name string, args ...string) *gocmd.Cmd {
	return executeInDir("", name, args...)
}

func executeInDir(dir, name string, args ...string) *gocmd.Cmd {
	command := gocmd.NewCmdOptions(cmd.StreamOutput, name, args...)
	command.Dir = dir
	done := command.Start()
	cmd.ShutdownHook(command, done)
	cmd.RedirectStreams(command, os.Stdout, os.Stderr, done)
	fmt.Printf("executing: [%s %v]\n", command.Name, command.Args)
	return command
}
