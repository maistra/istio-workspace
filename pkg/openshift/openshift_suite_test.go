package openshift_test

import (
	"testing"

	"go.uber.org/goleak"

	. "github.com/maistra/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOpenshift(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "Openshift object Suite")
}
