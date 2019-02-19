package watch_test

import (
	"github.com/aslakknutsen/istio-workspace/cmd/ike/watch"

	"github.com/fsnotify/fsnotify"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/aslakknutsen/istio-workspace/test"
)

var _ = Describe("File changes watch", func() {

	AfterEach(func() {
		CleanUp(GinkgoT())
	})

	It("should recognize file change", func() {
		// given
		done := make(chan struct{})
		config := TmpFile(GinkgoT(), "config.yaml", "content")

		watcher, e := watch.CreateWatch(1).
			WithHandler(expectFileChange(config.Name(), done)).
			OnPaths(config.Name())
		Expect(e).ToNot(HaveOccurred())

		defer watcher.Close()

		// when
		watcher.Watch()
		_, _ = config.WriteString(" modified!")

		// then
		Eventually(done).Should(BeClosed())
	})

	It("should recognize file change in watched directory", func() {
		// given
		done := make(chan struct{})
		tmpDir := TmpDir(GinkgoT(), "watch_yaml_txt")
		_ = TmpFile(GinkgoT(), tmpDir+"/config.yaml", "content")
		text := TmpFile(GinkgoT(), tmpDir+"/text.txt", "text text text")

		watcher, e := watch.CreateWatch(1).
			WithHandler(expectFileChange(text.Name(), done)).
			OnPaths(tmpDir)
		Expect(e).ToNot(HaveOccurred())

		defer watcher.Close()

		// when
		watcher.Watch()
		_, _ = text.WriteString(" modified!")

		// then
		Eventually(done).Should(BeClosed())
	})

	It("should recognize file change in sub-directory (recursive watch)", func() {
		// given
		done := make(chan struct{})
		tmpDir := TmpDir(GinkgoT(), "watch_yaml_txt")
		_ = TmpFile(GinkgoT(), tmpDir+"/config.yaml", "content")
		text := TmpFile(GinkgoT(), tmpDir+"/text.txt", "text text text")

		watcher, e := watch.CreateWatch(1).
			WithHandler(expectFileChange(text.Name(), done)).
			OnPaths(tmpDir)
		Expect(e).ToNot(HaveOccurred())

		defer watcher.Close()

		// when
		watcher.Watch()
		_, _ = text.WriteString(" modified!")

		// then
		Eventually(done).Should(BeClosed())
	})

	It("should not recognize file change if matches file extension exclusion", func() {
		// given
		done := make(chan struct{})
		tmpDir := TmpDir(GinkgoT(), "watch_yaml_txt")
		config := TmpFile(GinkgoT(), tmpDir+"/config.yaml", "content")
		text := TmpFile(GinkgoT(), tmpDir+"/text.txt", "text text text")

		watcher, e := watch.CreateWatch(1).
			WithHandler(expectFileChange(text.Name(), done)).
			Excluding("*.yaml").
			OnPaths(tmpDir)
		Expect(e).ToNot(HaveOccurred())

		defer watcher.Close()

		// when
		watcher.Watch()
		_, _ = config.WriteString(" should not be watched")
		_, _ = text.WriteString(" modified!")

		// then
		Eventually(done).Should(BeClosed())
	})

	It("should not recognize file change in excluded directory", func() {
		// given
		done := make(chan struct{})
		skipTmpDir := TmpDir(GinkgoT(), "skip_watch")

		config := TmpFile(GinkgoT(), skipTmpDir+"/config.yaml", "content")
		text := TmpFile(GinkgoT(), skipTmpDir+"/text.txt", "text text text")

		watchTmpDir := TmpDir(GinkgoT(), "watch")
		code := TmpFile(GinkgoT(), watchTmpDir+"/main.go", "package main")

		watcher, e := watch.CreateWatch(1).
			WithHandler(expectFileChange(code.Name(), done)).
			Excluding("/tmp/**/skip_watch/*").
			OnPaths(skipTmpDir, watchTmpDir)
		Expect(e).ToNot(HaveOccurred())

		defer watcher.Close()

		// when
		watcher.Watch()
		_, _ = config.WriteString(" should not be watched")
		_, _ = text.WriteString(" modified!")
		_, _ = code.WriteString("\n // Bla!")

		// then
		Eventually(done).Should(BeClosed())
	})

})

func expectFileChange(fileName string, done chan<- struct{}) watch.Handler {
	return func(events []fsnotify.Event) error {
		defer GinkgoRecover()
		Expect(events[0].Name).To(Equal(fileName))
		close(done)
		return nil
	}
}
