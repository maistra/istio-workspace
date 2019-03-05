package naming_test

import (
	"math/rand"
	"testing"
	"time"

	. "github.com/aslakknutsen/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNamingGenerator(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "Names Generator Suite")
}
