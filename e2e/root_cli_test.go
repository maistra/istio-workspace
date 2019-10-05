package e2e_test

import (
	"github.com/maistra/istio-workspace/test/shell"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Root CLI", func() {

	Context("exit codes", func() {

		It("should return non 0 on failed command", func() {
			completion := shell.ExecuteInDir(".", "bash", "-c", "ike missing-command")
			<-completion.Done()
			Expect(completion.Status().Exit).Should(Equal(23))
		})

	})
})
