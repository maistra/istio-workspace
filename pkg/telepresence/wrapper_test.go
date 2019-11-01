package telepresence_test

import (
	"os"
	"path"

	"github.com/maistra/istio-workspace/test/shell"

	"github.com/maistra/istio-workspace/pkg/telepresence"
	. "github.com/maistra/istio-workspace/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("telepresence commands wrapper", func() {

	Context("telepresence not available", func() {

		tmpPath := NewTmpPath()
		BeforeEach(func() {
			tmpPath.SetPath()
		})
		AfterEach(tmpPath.Restore)

		It("should fail when telepresence is not on $PATH", func() {
			Expect(telepresence.BinaryAvailable()).To(BeFalse())
		})

		It("should fail determining version when no env var nor telepresence binary available", func() {
			_, err := telepresence.GetVersion()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("couldn't find 'telepresence'"))
		})

	})

	Context("telepresence available", func() {

		It("should retrieve version from TELEPRESENCE_VERSION env variable when defined", func() {
			// given
			currentTpVersion := os.Getenv("TELEPRESENCE_VERSION")
			defer func() {
				if currentTpVersion != "" {
					_ = os.Setenv("TELEPRESENCE_VERSION", currentTpVersion)
				} else {
					_ = os.Unsetenv("TELEPRESENCE_VERSION")
				}
			}()
			_ = os.Setenv("TELEPRESENCE_VERSION", "0.123")

			// when
			version, err := telepresence.GetVersion()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("0.123"))
		})

		It("should retrieve version from telepresence binary", func() {
			tmpPath := NewTmpPath()
			tmpPath.SetPath(path.Dir(shell.TpVersionBin))
			defer tmpPath.Restore()

			// when
			version, err := telepresence.GetVersion()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("0.234"))
		})

	})

})
