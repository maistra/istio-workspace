package model_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/goleak"

	. "github.com/maistra/istio-workspace/test"
)

func TestModel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "model Suite")
}

var current goleak.Option

var _ = SynchronizedBeforeSuite(func() []byte {
	current = goleak.IgnoreCurrent()

	return []byte{}
}, func([]byte) {})

var _ = SynchronizedAfterSuite(func() {}, func() {
	goleak.VerifyNone(GinkgoT(), current)
})
