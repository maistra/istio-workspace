package watch_test

import (
	"path"

	"github.com/aslakknutsen/istio-workspace/cmd/ike/watch"

	"github.com/fsnotify/fsnotify"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/aslakknutsen/istio-workspace/test"
)

var _ = Describe("file changes watch", func() {

	AfterEach(func() {
		CleanUp(GinkgoT())
	})

	It("should watch on file change", func() {
		// given
		config := TmpFile(GinkgoT(), "config.yaml", "content")

		watcher, e := watch.NewWatch().
			WithHandler(expectFileChange(config.Name())).
			OnPaths(path.Dir(config.Name()))
		Expect(e).ToNot(HaveOccurred())

		defer watcher.Close()

		// when
		done := watcher.Watch()
		_, _ = config.WriteString("mod!")

		// then
		Eventually(done).Should(BeClosed())
	})

	It("should handle directory entry", func() {
		// given
		tmpDir := TmpDir(GinkgoT(), "watch_yaml_txt")
		_ = TmpFile(GinkgoT(), tmpDir+"/config.yaml", "content")
		text := TmpFile(GinkgoT(), tmpDir+"/text.txt", "text text text")

		watcher, e := watch.NewWatch().
			WithHandler(expectFileChange(text.Name())).
			OnPaths(tmpDir)
		Expect(e).ToNot(HaveOccurred())

		defer watcher.Close()

		// when
		done := watcher.Watch()
		_, _ = text.WriteString("mod!")

		// then
		Eventually(done).Should(BeClosed())
	})

	It("should watch directories recursively", func() {
		// given
		tmpDir := TmpDir(GinkgoT(), "watch_yaml_txt")
		_ = TmpFile(GinkgoT(), tmpDir+"/config.yaml", "content")
		text := TmpFile(GinkgoT(), tmpDir+"/text.txt", "text text text")

		watcher, e := watch.NewWatch().
			WithHandler(expectFileChange(text.Name())).
			OnPaths(tmpDir)
		Expect(e).ToNot(HaveOccurred())

		defer watcher.Close()

		// when
		done := watcher.Watch()
		_, _ = text.WriteString(" modified!")

		// then
		Eventually(done).Should(BeClosed())
	})

	It("should skip watching files when file exclusion pattern matches", func() {
		// given
		tmpDir := TmpDir(GinkgoT(), "watch_yaml_txt")
		config := TmpFile(GinkgoT(), tmpDir+"/config.yaml", "content")
		text := TmpFile(GinkgoT(), tmpDir+"/text.txt", "text text text")

		watcher, e := watch.NewWatch().
			WithHandler(expectFileChange(text.Name())).
			Excluding("*.yaml").
			OnPaths(tmpDir)
		Expect(e).ToNot(HaveOccurred())

		defer watcher.Close()

		// when
		done := watcher.Watch()
		_, _ = config.WriteString(" should not be watched")
		_, _ = text.WriteString(" modified!")

		// then
		Eventually(done).Should(BeClosed())
	})

	It("should skip watching files when directory exclusion pattern matches", func() {
		// given
		skipTmpDir := TmpDir(GinkgoT(), "skip_watch")

		config := TmpFile(GinkgoT(), skipTmpDir+"/config.yaml", "content")
		text := TmpFile(GinkgoT(), skipTmpDir+"/text.txt", "text text text")

		watchTmpDir := TmpDir(GinkgoT(), "watch")
		code := TmpFile(GinkgoT(), watchTmpDir+"/main.go", "package main")

		watcher, e := watch.NewWatch().
			WithHandler(expectFileChange(code.Name())).
			Excluding("/tmp/**/skip_watch/*").
			OnPaths(skipTmpDir, watchTmpDir)
		Expect(e).ToNot(HaveOccurred())

		defer watcher.Close()

		// when
		done := watcher.Watch()
		_, _ = config.WriteString(" should not be watched")
		_, _ = text.WriteString(" modified!")
		_, _ = code.WriteString("\n // Bla!")

		// then
		Eventually(done).Should(BeClosed())
	})

})

func expectFileChange(fileName string) watch.Handler {
	return func(event fsnotify.Event, done chan<- struct{}) error {
		defer GinkgoRecover()
		close(done)
		Expect(event.Name).To(Equal(fileName))
		return nil
	}
}
