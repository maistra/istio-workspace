package version_test

import (
	"gopkg.in/h2non/gock.v1"

	"github.com/maistra/istio-workspace/pkg/cmd/version"
	v "github.com/maistra/istio-workspace/version"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Usage of ike develop command", func() {

	BeforeEach(func() {
		gock.New("https://api.github.com").
			Get("/repos/maistra/istio-workspace/releases/latest").
			Reply(200).
			File("fixtures/latest_release_is_v.0.0.2.json")
	})

	AfterEach(func() {
		gock.Off()
	})

	It("should get latest release", func() {
		// when
		release, err := version.LatestRelease()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(release).To(Equal("v0.0.2"))
	})

	It("should determine that v0.0.0 is not latest release", func() {
		// given
		latestRelease, err := version.LatestRelease()
		Expect(err).ToNot(HaveOccurred())

		// when
		latest := version.IsLatestRelease(latestRelease)

		// then
		Expect(latest).To(BeFalse())
	})

	It("should determine that v0.0.2 is latest release", func() {
		// given
		v.Version = "v0.0.2"
		latestRelease, err := version.LatestRelease()
		Expect(err).ToNot(HaveOccurred())

		// when
		latest := version.IsLatestRelease(latestRelease)

		// then
		Expect(latest).To(BeTrue())
	})

})
