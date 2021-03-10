package create_test

import (
	. "github.com/maistra/istio-workspace/pkg/cmd"
	"github.com/maistra/istio-workspace/pkg/cmd/create"
	. "github.com/maistra/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"
)

var _ = Describe("Usage of ike create command", func() {

	var createCmd *cobra.Command

	BeforeEach(func() {
		createCmd = create.NewCmd()
		createCmd.SilenceUsage = true
		createCmd.SilenceErrors = true
		NewCmd().AddCommand(createCmd)
	})

	Describe("input validation", func() {

		Context("with flags only", func() {

			It("should fail when deployment is not specified", func() {
				_, err := ValidateArgumentsOf(createCmd).Passing("--image", "1234")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(And(ContainSubstring("required flag(s)"), ContainSubstring("deployment")))
			})

			It("should fail when image command is not specified", func() {
				defer TemporaryUnsetEnvVars("IKE_IMAGE")()
				_, err := ValidateArgumentsOf(createCmd).Passing("--deployment", "rating-service")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(And(ContainSubstring("required flag(s)"), ContainSubstring("image")))
			})

			It("should be able to provide the traffic route parameter", func() {
				_, err := ValidateArgumentsOf(createCmd).Passing("--deployment", "rating-service", "--image", "x", "--route", "header:name=value")

				Expect(err).NotTo(HaveOccurred())
				Expect(createCmd.Flag("route").Value.String()).To(Equal("header:name=value"))
			})

		})

	})

})
