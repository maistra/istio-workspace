package verify

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	. "github.com/maistra/istio-workspace/e2e/infra"
)

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
