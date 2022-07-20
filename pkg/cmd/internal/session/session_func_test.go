package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/maistra/istio-workspace/pkg/cmd/develop"
	internal "github.com/maistra/istio-workspace/pkg/cmd/internal/session"
	"github.com/maistra/istio-workspace/test"
)

var _ = Describe("Usage of session func", func() {

	var restoreEnvVars func()

	JustBeforeEach(func() {
		restoreEnvVars = test.TemporaryEnvVars("TELEPRESENCE_VERSION", "0.123")
	})

	JustAfterEach(func() {
		restoreEnvVars()
	})

	Context("checking required command arguments", func() {

		var command *cobra.Command

		BeforeEach(func() {
			command = develop.NewCmd()
		})

		It("should fail if namespace is not defined", func() {
			_, err := internal.ToOptions(command.Annotations, removeFlagFromSet(command, "namespace"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("namespace"))
		})

		It("should fail if deployment is not defined", func() {
			_, err := internal.ToOptions(command.Annotations, removeFlagFromSet(command, "deployment"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("deployment"))
		})

		It("should fail if session is not defined", func() {
			_, err := internal.ToOptions(command.Annotations, removeFlagFromSet(command, "session"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("session"))
		})

		It("should fail if route is not defined", func() {
			_, err := internal.ToOptions(command.Annotations, removeFlagFromSet(command, "route"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("route"))
		})
	})

	Context("checking conversion", func() {

		var command *cobra.Command

		BeforeEach(func() {
			command = develop.NewCmd()
		})

		It("should convert namespace if set", func() {
			Expect(command.Flag("namespace").Value.Set("TEST")).ToNot(HaveOccurred())
			opts, err := internal.ToOptions(command.Annotations, internal.CollectFlags(command))
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.NamespaceName).To(Equal("TEST"))
		})

		It("should convert deployment if set", func() {
			Expect(command.Flag("deployment").Value.Set("TEST")).ToNot(HaveOccurred())
			opts, err := internal.ToOptions(command.Annotations, internal.CollectFlags(command))
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.DeploymentName).To(Equal("TEST"))
		})

		It("should convert session if set", func() {
			Expect(command.Flag("session").Value.Set("TEST")).ToNot(HaveOccurred())
			opts, err := internal.ToOptions(command.Annotations, internal.CollectFlags(command))
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.SessionName).To(Equal("TEST"))
		})

		It("should convert route if set", func() {
			// RouteExp Parser not tested here, see session/session_test
			Expect(command.Flag("route").Value.Set("header:name=value")).ToNot(HaveOccurred())
			opts, err := internal.ToOptions(command.Annotations, internal.CollectFlags(command))
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.RouteExp).To(Equal("header:name=value"))
		})

		It("should set Revert if command is develop", func() {
			opts, err := internal.ToOptions(command.Annotations, internal.CollectFlags(command))
			Expect(err).ToNot(HaveOccurred())
			Expect(opts.Revert).To(BeTrue())
		})

		It("should set Revert if command has revert annotation", func() {
			annotations := map[string]string{
				internal.AnnotationRevert: "true",
			}
			opts, err := internal.ToOptions(annotations, internal.CollectFlags(command))
			Expect(err).ToNot(HaveOccurred())
			Expect(opts.Revert).To(BeTrue())
		})

		It("should not set Revert if command is not develop", func() {
			annotations := map[string]string{
				internal.AnnotationRevert: "false",
			}
			opts, err := internal.ToOptions(annotations, internal.CollectFlags(command))
			Expect(err).ToNot(HaveOccurred())
			Expect(opts.Revert).To(BeFalse())
		})

		It("should default to empty", func() {
			opts, err := internal.ToOptions(command.Annotations, internal.CollectFlags(command))
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.NamespaceName).To(Equal(""))
			Expect(opts.DeploymentName).To(Equal(""))
			Expect(opts.SessionName).To(Equal(""))
			Expect(opts.RouteExp).To(Equal(""))
		})

	})
})

func removeFlagFromSet(cmd *cobra.Command, flagToRemove string) map[string]string {
	f := pflag.NewFlagSet("combined-flags", pflag.ContinueOnError)
	copyFlags := func(flag *pflag.Flag) {
		if flag.Name != flagToRemove {
			f.AddFlag(flag)
		}
	}
	cmd.Flags().VisitAll(copyFlags)
	cmd.PersistentFlags().VisitAll(copyFlags)

	flags := map[string]string{}
	collect := func(flag *pflag.Flag) {
		flags[flag.Name] = flag.Value.String()
	}
	f.VisitAll(collect)
	f.VisitAll(collect)

	return flags
}
