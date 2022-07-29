package flag_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	. "github.com/maistra/istio-workspace/pkg/cmd"
	"github.com/maistra/istio-workspace/pkg/cmd/flag"
	. "github.com/maistra/istio-workspace/test"
)

var _ = Describe("Usage of limited flags", func() {

	Context("when passing flags explicitly", func() {

		var testCmd *cobra.Command

		BeforeEach(func() {
			testCmd = newTestCmd("stout", "s", "ale", "a", "kolsch", "k")
			testCmd.SilenceUsage = true
			testCmd.SilenceErrors = true
			NewCmd().AddCommand(testCmd)
		})

		It("should accept defined value using full name", func() {
			output, err := Run(testCmd).Passing("--style", "stout")
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("selected style: 'stout'"))
		})

		It("should accept defined value using abbreviated name", func() {
			output, err := Run(testCmd).Passing("--style", "k")
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("selected style: 'kolsch'"))
		})

		It("should fail when wrong argument passed", func() {
			_, err := Run(testCmd).Passing("--style", "neipa")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`invalid argument "neipa" for "--style" flag: must be one of [stout (s) ale (a) kolsch (k)]`))
		})

	})

	// Autocompletion is covered in e2e-tests

})

func newTestCmd(namesAndAbbrevs ...string) *cobra.Command {
	testCmd := &cobra.Command{
		Use:          "test",
		Short:        "Test command",
		SilenceUsage: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			value := cmd.Flag("style").Value.String()
			cmd.Printf("selected style: '%s'", value)

			return nil
		},
	}

	beerStyles := flag.CreateOptions(namesAndAbbrevs...)
	beerStyle := beerStyles[0]
	testCmd.Flags().Var(&beerStyle, "style", "beer styles")
	_ = testCmd.RegisterFlagCompletionFunc("style", flag.CompletionFor(beerStyles))

	return testCmd
}
