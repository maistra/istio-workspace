package watch_test

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/maistra/istio-workspace/pkg/watch"

	"github.com/fsnotify/fsnotify"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maistra/istio-workspace/test"
)

var _ = Describe("File changes watch", func() {

	It("should recognize file change", func() {
		// given
		done := make(chan struct{})
		config := TmpFile(GinkgoT(), "config.yaml", "content")

		watcher, e := watch.CreateWatch(1).
			WithHandlers(expectFileChange(config.Name(), done)).
			OnPaths(config.Name())
		Expect(e).ToNot(HaveOccurred())

		defer watcher.Close()

		// when
		watcher.Start()
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
			WithHandlers(expectFileChange(text.Name(), done)).
			OnPaths(tmpDir)
		Expect(e).ToNot(HaveOccurred())

		defer watcher.Close()

		// when
		watcher.Start()
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
			WithHandlers(expectFileChange(text.Name(), done)).
			OnPaths(tmpDir)
		Expect(e).ToNot(HaveOccurred())

		defer watcher.Close()

		// when
		watcher.Start()
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
			WithHandlers(expectFileChange(text.Name(), done), notExpectFileChange(config.Name())).
			Excluding("*.yaml").
			OnPaths(tmpDir)
		Expect(e).ToNot(HaveOccurred())

		defer watcher.Close()

		// when
		watcher.Start()
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

		watchTmpDir := TmpDir(GinkgoT(), "watch")
		code := TmpFile(GinkgoT(), watchTmpDir+"/main.go", "package main")

		watcher, e := watch.CreateWatch(1).
			WithHandlers(notExpectFileChange(config.Name()), expectFileChange(code.Name(), done)).
			Excluding("skip_watch/**").
			OnPaths(filepath.Dir(skipTmpDir), filepath.Dir(watchTmpDir))
		Expect(e).ToNot(HaveOccurred())

		defer watcher.Close()

		// when
		watcher.Start()
		_, _ = config.WriteString(" should not be watched")
		_, _ = code.WriteString("\n // Bla!")

		// then
		Eventually(done).Should(BeClosed())
	})

	It("should not recognize file change if it's git-ignored", func() {
		// given
		done := make(chan struct{})

		watchTmpDir := TmpDir(GinkgoT(), "watch")
		config := TmpFile(GinkgoT(), watchTmpDir+"/.idea/config.toml", "content")
		test := TmpFile(GinkgoT(), watchTmpDir+"/.idea/test.yaml", "content")
		nestedFile := TmpFile(GinkgoT(), watchTmpDir+"/src/main/org/acme/Main.java", "package org.acme")
		gitIgnore := TmpFile(GinkgoT(), watchTmpDir+"/.gitignore", `
*.yaml
src/main/**/*.java
.idea/
`)
		code := TmpFile(GinkgoT(), watchTmpDir+"/main.go", "package main")

		defer func() {
			if err := config.Close(); err != nil {
				fmt.Println(err)
			}
			if err := test.Close(); err != nil {
				fmt.Println(err)
			}
			if err := nestedFile.Close(); err != nil {
				fmt.Println(err)
			}
			if err := gitIgnore.Close(); err != nil {
				fmt.Println(err)
			}
		}()

		watcher, e := watch.CreateWatch(1).
			WithHandlers(notExpectFileChange(config.Name(), test.Name(), nestedFile.Name()), expectFileChange(code.Name(), done)).
			OnPaths(watchTmpDir)
		Expect(e).ToNot(HaveOccurred())

		defer watcher.Close()

		// when
		watcher.Start()
		_, _ = test.WriteString(" should not be watched")
		_, _ = config.WriteString(" should not be watched")
		_, _ = nestedFile.WriteString("\n// TODO implement me but now ignore")
		_, _ = code.WriteString("\n // Bla! Should trigger watch reaction")

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

func notExpectFileChange(fileNames ...string) watch.Handler {
	return func(events []fsnotify.Event) error {
		defer GinkgoRecover()
		for _, event := range events {
			for _, fileName := range fileNames {
				if event.Name == fileName {
					errMsg := fmt.Sprintf("expected %s to not change", fileName)
					Fail(errMsg)
					return errors.New(errMsg)
				}
			}
		}
		return nil
	}
}
