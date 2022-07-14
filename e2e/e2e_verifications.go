package e2e

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"emperror.dev/errors"
	"github.com/go-cmd/cmd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/schollz/progressbar/v3"

	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/pkg/naming"
	"github.com/maistra/istio-workspace/test"
	testshell "github.com/maistra/istio-workspace/test/shell"
)

// EnsureAllDeploymentPodsAreReady make sure all Pods are in Ready state in given namespace.
func EnsureAllDeploymentPodsAreReady(namespace string) {
	Eventually(AllDeploymentsAndPodsReady(namespace), 5*time.Minute, 5*time.Second).Should(BeTrue())
}

// EnsureAllDeploymentConfigPodsAreReady make sure all Pods are in Ready state in given namespace.
func EnsureAllDeploymentConfigPodsAreReady(namespace string) {
	Eventually(AllDeploymentConfigsAndPodsReady(namespace), 10*time.Minute, 5*time.Second).Should(BeTrue())
}

// EnsureCorrectNumberOfResources make sure the correct number of given resource are in namespace.
func EnsureCorrectNumberOfResources(count int, resource, namespace string) {
	Eventually(MatchResourceCount(count, GetResourceCountFunc(resource, namespace)), 5*time.Minute, 5*time.Second).Should(BeTrue())
}

// EnsureProdRouteIsReachable can be reached with no special arguments.
func EnsureProdRouteIsReachable(namespace string, matchers ...types.GomegaMatcher) {
	productPageURL := GetIstioIngressHostname() + "/productpage"

	Eventually(call(productPageURL, map[string]string{
		"Host": GetGatewayHost(namespace)}),
		10*time.Minute, 1*time.Second).Should(And(matchers...))
}

type stableCountMatcher struct {
	delegate              types.GomegaMatcher
	matchCount            int32
	subsequentOccurrences int32
	flipping              bool
}

func (s *stableCountMatcher) Match(actual interface{}) (success bool, err error) {
	match, err := s.delegate.Match(actual)
	if !match {
		if s.matchCount > 0 {
			s.flipping = true
		}
		s.matchCount = 0

		return false, err
	}

	s.matchCount++

	if s.matchCount < s.subsequentOccurrences {
		return false, errors.Errorf("not enough matches in sequence yet [%d/%d]", s.matchCount, s.subsequentOccurrences)
	}

	return match, err
}

func (s *stableCountMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf(
		"failed to receive stable response after %d times. Response is flipping:%v. latest cause: %s",
		s.subsequentOccurrences, s.flipping, s.delegate.FailureMessage(actual))
}

func (s *stableCountMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf(
		"failed to receive stable response after %d times. Response is flipping:%v. latest cause: %s",
		s.subsequentOccurrences, s.flipping, s.delegate.NegatedFailureMessage(actual))
}

func beStableInSeries(occurrences int32, matcher types.GomegaMatcher) types.GomegaMatcher {
	return &stableCountMatcher{subsequentOccurrences: occurrences, delegate: matcher}
}

// EnsureSessionRouteIsReachable the manipulated route is reachable.
func EnsureSessionRouteIsReachable(namespace, sessionName string, matchers ...types.GomegaMatcher) {
	productPageURL := GetIstioIngressHostname() + "/productpage"

	By("checking response using headers")
	Eventually(call(productPageURL, map[string]string{
		"Host":         GetGatewayHost(namespace),
		"x-test-suite": "smoke"}),
		10*time.Minute, 1*time.Second).Should(beStableInSeries(8, And(matchers...)))

	By("checking response using host")
	Eventually(call(productPageURL, map[string]string{
		"Host": sessionName + "." + GetGatewayHost(namespace)}),
		10*time.Minute, 1*time.Second).Should(beStableInSeries(8, And(matchers...)))
}

// EnsureSessionRouteIsNotReachable the manipulated route is reachable.
func EnsureSessionRouteIsNotReachable(namespace, sessionName string, matchers ...types.GomegaMatcher) {
	productPageURL := GetIstioIngressHostname() + "/productpage"

	// check original response using headers
	Eventually(call(productPageURL, map[string]string{
		"Host":         GetGatewayHost(namespace),
		"x-test-suite": "smoke"}),
		10*time.Minute, 1*time.Second).Should(And(matchers...))
}

// ChangeNamespace switch to different namespace - so we also test -n parameter of $ ike.
// That only works for oc cli, as kubectl by default uses `default` namespace.
func ChangeNamespace(namespace string) {
	if RunsOnOpenshift {
		<-testshell.Execute("oc project " + namespace).Done()
	}
}

// RunIke runs the ike cli in the given dir.
func RunIke(dir string, arguments ...string) *cmd.Cmd {
	return testshell.ExecuteInDir(dir, "ike", arguments...)
}

// Stop shuts down the process.
func Stop(ike *cmd.Cmd) {
	stopFailed := ike.Stop()
	Expect(stopFailed).ToNot(HaveOccurred())

	Eventually(ike.Done(), 1*time.Minute).Should(BeClosed())
}

func FailOnCmdError(command *cmd.Cmd, t test.TestReporter) {
	<-command.Done()
	if command.Status().Exit != 0 {
		t.Errorf("failed executing %s with code %d", command.Name, command.Status().Exit)
	}
}

// DumpEnvironmentDebugInfo prints tons of noise about the cluster state when test fails.
func DumpEnvironmentDebugInfo(namespace, dir string) {
	GetEvents(namespace)
	DumpTelepresenceLog(dir)
}

func GenerateNamespaceName() string {
	return "ike-tests-" + naming.GenerateString(16)
}

func CleanupNamespace(namespace string, wait bool) {
	if keepStr, found := os.LookupEnv("IKE_E2E_KEEP_NS"); found {
		keep, _ := strconv.ParseBool(keepStr)
		if keep {
			return
		}
	}
	CleanupTestScenario(namespace)
	<-testshell.Execute("kubectl delete namespace " + namespace + " --wait=" + strconv.FormatBool(wait)).Done()
}

func call(routeURL string, headers map[string]string) func() (string, error) {
	fmt.Printf("Checking [%s] with headers [%s]\n", routeURL, headers)
	bar := progressbar.Default(-1)

	return func() (string, error) {
		bar.Add(1)

		return GetBodyWithHeaders(routeURL, headers)
	}
}
