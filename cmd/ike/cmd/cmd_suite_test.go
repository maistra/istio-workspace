package cmd_test

import (
	. "github.com/aslakknutsen/istio-workspace/test"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "CLI Suite")
}
