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

var _ = SynchronizedBeforeSuite(func() []byte {

	Expect(cmd.BinaryExists("istiooc", "check https://maistra.io/ for details")).To(BeTrue())
	Expect(cmd.BinaryExists("oc", "grab latest openshift origin client tools from here https://github.com/openshift/origin/releases")).To(BeTrue())
	Expect(cmd.BinaryExists("python3", "make sure you have python3 installed")).To(BeTrue())

	measure(func() {
		fmt.Println("\nStarting up Openshift/Istio cluster")
		<-execute("istiooc", "cluster", "up",
			// TODO tmp folder - but probably not to clean up
			"--base-dir", "/home/bartek/code/clusters/openshift/istio-workspace-maistra/mycluster-ocp").Done()
	})

	return nil
},
	func(data []byte) {})

var _ = SynchronizedAfterSuite(func() {},
	func() {
		measure(func() {
			fmt.Println("\nStopping Openshift/Istio cluster")
			<-execute("oc", "cluster", "down").Done()
		})
	})

type noArgFunc func()

func measure(f noArgFunc) {
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
