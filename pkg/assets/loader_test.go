package assets //nolint:testpackage //reason we want to stub go-bindata internal structs

import (
	"os"
	"time"

	"github.com/maistra/istio-workspace/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Loading assets", func() {

	var tmpDirPath string
	tmpDirName := "loader-tests"
	tmpFs := test.NewTmpFileSystem(GinkgoT())

	assetName := "test-file-001.yaml"

	JustAfterEach(func() {
		tmpFs.Cleanup()
	})

	Context("Listing assets", func() {
		Context("Local file system", func() {
			It("should list files from existing directory", func() {
				// given
				tmpDirPath = tmpFs.Dir(tmpDirName)

				fileName := tmpDirPath + "/" + assetName
				tmpFs.File(fileName, "it works!")

				// when
				files, err := ListDir(tmpDirPath)
				Expect(err).NotTo(HaveOccurred())

				// then
				Expect(files).To(ConsistOf("test-file-001.yaml"))
			})

			It("should return empty list for empty dir", func() {
				// given
				tmpDirPath = tmpFs.Dir("loader-tests")

				// when
				files, err := ListDir(tmpDirPath)
				Expect(err).NotTo(HaveOccurred())

				// then
				Expect(files).To(HaveLen(0))
			})
		})
	})

	Context("built-in assets", func() {

		It("list existing assets", func() {
			// given
			fileName := tmpDirName + "/" + assetName
			stubBinDataStructs(tmpDirName, assetName, fileName)

			// when
			dir, err := ListDir(tmpDirName) // as test is run in its own working directory, not project's root
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(dir).To(HaveLen(1))
		})

		// Empty folders are not added to generated code when running go-bindata therefore empty asset check is not a thing
	})

	Context("Loading assets", func() {

		Context("both filesystem and generated in-memory assets have the same entry", func() {

			It("should load file from filesystem", func() {
				// given
				tmpDirPath = tmpFs.Dir(tmpDirName)

				fileName := tmpDirPath + "/" + assetName
				tmpFs.File(fileName, "it works!")

				stubBinDataStructs(tmpDirName, assetName, fileName)

				// when
				content, err := Load(fileName)
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(string(content)).To(Equal("it works!"))
			})
		})

		Context("from in-memory", func() {
			It("should load existing asset", func() {
				// given
				fileName := tmpDirName + "/" + assetName
				stubBinDataStructs(tmpDirName, assetName, fileName)

				// when
				content, err := Load(fileName)
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(string(content)).To(Equal("it works also here\n"))
			})
		})

	})

})

func stubBinDataStructs(tmpDirName, assetName, fileName string) {
	_bintree = &bintree{nil, map[string]*bintree{
		tmpDirName: {nil, map[string]*bintree{
			assetName: {testYaml(fileName), map[string]*bintree{}},
		}},
	}}

	_bindata = map[string]func() (*asset, error){
		fileName: testYaml(fileName),
	}
}

// Stubbed asset generation

var _testYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xca\x2c\x51\x28\xcf\x2f\xca\x2e\x56\x48\xcc\x29\xce\x57\xc8" +
	"\x48\x2d\x4a\xe5\x02\x04\x00\x00\xff\xff\x26\xff\x91\x56\x13\x00\x00\x00")

func testYamlBytes(name string) ([]byte, error) {
	return bindataRead(
		_testYaml,
		name,
	)
}

func testYaml(name string) func() (*asset, error) {
	return func() (*asset, error) {
		bytes, err := testYamlBytes(name)
		if err != nil {
			return nil, err
		}

		info := bindataFileInfo{name: name, size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
		a := &asset{bytes: bytes, info: info}

		return a, nil
	}
}
