package naming_test

import (
	"math/rand"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/maistra/istio-workspace/pkg/naming"
	. "github.com/maistra/istio-workspace/test/matchers"
)

var _ = Describe("Name generation (used for k8s objects such as namespaces, sessions etc)", func() {

	It("should always generate lowercase string", func() {
		name := naming.GenerateString(32)
		Expect(name).To(Equal(strings.ToLower(name)))
	})

	It("should generate empty string when 0 length requested", func() {
		name := naming.GenerateString(0)
		Expect(name).To(BeEmpty())
	})

	It("should always generate letter when single character name is requested", func() {
		name := naming.GenerateString(1)
		Expect(name).To(OnlyContain("abcdefghijklmnopqrstuvwxyz"))
	})

	It("should generate name only with letters", func() {
		name := naming.GenerateString(rand.Intn(512) + 2)
		Expect(name).To(OnlyContain("abcdefghijklmnopqrstuvwxyz"))
	})

	It("should trim length to 63 when exceeded ", func() {
		name := naming.GenerateString(rand.Intn(512) + 59)
		Expect(name).To(HaveLen(63))
	})

})
