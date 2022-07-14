package verify

import (
	"fmt"
	"time"

	"emperror.dev/errors"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/maistra/istio-workspace/e2e/infra"
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
	productPageURL := infra.GetIstioIngressHostname() + "/productpage"

	Eventually(call(productPageURL, map[string]string{
		"Host": infra.GetGatewayHost(namespace)}),
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
