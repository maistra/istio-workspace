package watch_test

import (
	"testing"

	. "github.com/aslakknutsen/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWatch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "Watch Suite")
}
