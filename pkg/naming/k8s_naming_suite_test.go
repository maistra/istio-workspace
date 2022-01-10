package naming_test

import (
	"math/rand"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/goleak"
)

func TestNamingGenerator(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	RegisterFailHandler(Fail)
	RunSpecs(t, "Names Generator Suite")
}

var current goleak.Option

var _ = SynchronizedBeforeSuite(func() []byte {
	current = goleak.IgnoreCurrent()

	return []byte{}
}, func([]byte) {})

var _ = SynchronizedAfterSuite(func() {}, func() {
	goleak.VerifyNone(GinkgoT(), current)
})
