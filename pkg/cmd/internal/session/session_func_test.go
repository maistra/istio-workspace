package internal_test

import (
	"github.com/maistra/istio-workspace/pkg/cmd/develop"
	internal "github.com/maistra/istio-workspace/pkg/cmd/internal/session"
	"github.com/maistra/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var _ = Describe("Usage of session func", func() {

	envVars := test.TemporaryEnvVars()

	BeforeEach(func() {
		envVars.Set("TELEPRESENCE_VERSION", "0.123")
	})

	AfterEach(envVars.Restore)

	Context("checking required command arguments", func() {

		var command *cobra.Command

		BeforeEach(func() {
			command = develop.NewCmd()
		})

		It("should fail if namespace is not defined", func() {
			_, err := internal.ToOptions(removeFlagFromSet(command.Flags(), "namespace"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("namespace"))
		})

		It("should fail if deployment is not defined", func() {
			_, err := internal.ToOptions(removeFlagFromSet(command.Flags(), "deployment"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("deployment"))
		})

		It("should fail if session is not defined", func() {
			_, err := internal.ToOptions(removeFlagFromSet(command.Flags(), "session"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("session"))
		})

		It("should fail if route is not defined", func() {
			_, err := internal.ToOptions(removeFlagFromSet(command.Flags(), "route"))
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
			Expect(command.Flags().Set("namespace", "TEST")).ToNot(HaveOccurred())
			opts, err := internal.ToOptions(command.Flags())
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.NamespaceName).To(Equal("TEST"))
		})

		It("should convert deployment if set", func() {
			Expect(command.Flags().Set("deployment", "TEST")).ToNot(HaveOccurred())
			opts, err := internal.ToOptions(command.Flags())
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.DeploymentName).To(Equal("TEST"))
		})

		It("should convert session if set", func() {
			Expect(command.Flags().Set("session", "TEST")).ToNot(HaveOccurred())
			opts, err := internal.ToOptions(command.Flags())
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.SessionName).To(Equal("TEST"))
		})

		It("should convert route if set", func() {
			// RouteExp Parser not tested here, see session/session_test
			Expect(command.Flags().Set("route", "header:name=value")).ToNot(HaveOccurred())
			opts, err := internal.ToOptions(command.Flags())
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.RouteExp).To(Equal("header:name=value"))
		})

		It("should default to empty", func() {
			opts, err := internal.ToOptions(command.Flags())
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.NamespaceName).To(Equal(""))
			Expect(opts.DeploymentName).To(Equal(""))
			Expect(opts.SessionName).To(Equal(""))
			Expect(opts.RouteExp).To(Equal(""))
		})

	})
})

func removeFlagFromSet(flags *pflag.FlagSet, flagToRemove string) *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Name != flagToRemove {
			f.AddFlag(flag)
		}
	})
	return f
}
