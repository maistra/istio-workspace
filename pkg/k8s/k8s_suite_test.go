package k8s_test

import (
	"testing"

	. "github.com/aslakknutsen/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestK8(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "K8 object Suite")
}
