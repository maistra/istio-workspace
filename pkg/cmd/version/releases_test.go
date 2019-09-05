package version_test

import (
	"gopkg.in/h2non/gock.v1"

	"github.com/maistra/istio-workspace/pkg/cmd/version"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Usage of ike develop command", func() {

	It("should get latest release", func() {
		// given
		defer gock.Off()

		gock.New("https://api.github.com").
			Get("/repos/maistra/istio-workspace/releases/latest").
			Reply(200).
			File("fixtures/latest_release.json")

		// when
		release, err := version.LatestRelease()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(release).To(Equal("v0.0.2"))
	})

})
