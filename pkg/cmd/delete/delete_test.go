package delete_test

import (
	"path"

	"github.com/maistra/istio-workspace/test/shell"

	. "github.com/maistra/istio-workspace/pkg/cmd"
	"github.com/maistra/istio-workspace/pkg/cmd/delete"
	. "github.com/maistra/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"
)

var _ = Describe("Usage of ike create command", func() {

	var createCmd *cobra.Command

	BeforeEach(func() {
		createCmd = delete.NewCmd()
		createCmd.SilenceUsage = true
		createCmd.SilenceErrors = true
		NewCmd().AddCommand(createCmd)
	})

	Describe("input validation", func() {

		tmpPath := NewTmpPath()
		BeforeEach(func() {
			tmpPath.SetPath(path.Dir(shell.MvnBin), path.Dir(shell.TpSleepBin))
		})
		AfterEach(tmpPath.Restore)

		Context("with flags only", func() {

			It("should fail when deployment is not specified", func() {
				_, err := ValidateArgumentsOf(createCmd).Passing("-s x --namespace", "1234")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(And(ContainSubstring("required flag(s)"), ContainSubstring("deployment")))
			})

			It("should fail when session is not specified", func() {
				_, err := ValidateArgumentsOf(createCmd).Passing("-d x --namespace", "1234")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(And(ContainSubstring("required flag(s)"), ContainSubstring("session")))
			})

		})

	})

})
