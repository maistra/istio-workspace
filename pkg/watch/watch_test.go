package watch_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"emperror.dev/errors"
	"github.com/fsnotify/fsnotify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/maistra/istio-workspace/pkg/watch"
	. "github.com/maistra/istio-workspace/test"
)

const pathSeparator = string(os.PathSeparator)

var _ = Describe("Watching for file changes", func() {

	tmpFs := NewTmpFileSystem(GinkgoT())

	AfterEach(func() {
		tmpFs.Cleanup()
	})

	Context("in the current directory", func() {

		It("should react on file change for the file starting with '.'", func() {
			// given
			done := make(chan struct{})

			_, testFile, _, _ := runtime.Caller(0)
			fullPath := filepath.Dir(testFile) + string(os.PathSeparator) + ".run.exe" // Adding .exe as it's gitignored for the repo
			fileToWatch := tmpFs.File(fullPath, "content")

			watcher, e := watch.CreateWatch(1).
				WithHandlers(expectChangesOf(fileToWatch.Name(), done)).
				OnPaths(".") // current directory
			Expect(e).ToNot(HaveOccurred())

			defer watcher.Close()

			// when
			watcher.Start()
			_, _ = fileToWatch.WriteString(" modified!")

			// then
			Eventually(done).Should(BeClosed())
		})
	})

	Context("in the arbitrary directory", func() {

		It("should react on file change", func() {
			// given
			done := make(chan struct{})
			config := tmpFs.File("config.yaml", "content")

			watcher, e := watch.CreateWatch(1).
				WithHandlers(expectChangesOf(config.Name(), done)).
				OnPaths(config.Name())
			Expect(e).ToNot(HaveOccurred())

			defer watcher.Close()

			// when
			watcher.Start()
			_, _ = config.WriteString(" modified!")

			// then
			Eventually(done).Should(BeClosed())
		})

		It("should react on file change in watched directory", func() {
			// given
			done := make(chan struct{})
			tmpDir := tmpFs.Dir("watch_yaml_txt")
			_ = tmpFs.File(tmpDir+pathSeparator+"config.yaml", "content")
			text := tmpFs.File(tmpDir+pathSeparator+"text.txt", "text text text")

			watcher, e := watch.CreateWatch(1).
				WithHandlers(expectChangesOf(text.Name(), done)).
				OnPaths(tmpDir)
			Expect(e).ToNot(HaveOccurred())

			defer watcher.Close()

			// when
			watcher.Start()
			_, _ = text.WriteString(" modified!")

			// then
			Eventually(done).Should(BeClosed())
		})

		It("should react on file change in sub-directory (recursive watch)", func() {
			// given
			done := make(chan struct{})
			tmpDir := tmpFs.Dir("watch_yaml_txt")
			_ = tmpFs.File(tmpDir+pathSeparator+"config.yaml", "content")
			text := tmpFs.File(tmpDir+pathSeparator+"text.txt", "text text text")

			watcher, e := watch.CreateWatch(1).
				WithHandlers(expectChangesOf(text.Name(), done)).
				OnPaths(tmpDir)
			Expect(e).ToNot(HaveOccurred())

			defer watcher.Close()

			// when
			watcher.Start()
			_, _ = text.WriteString(" modified!")

			// then
			Eventually(done).Should(BeClosed())
		})

		It("should not react on file change if matches file extension exclusion", func() {
			// given
			done := make(chan struct{})
			tmpDir := tmpFs.Dir("watch_yaml_txt")
			config := tmpFs.File(tmpDir+pathSeparator+"config.yaml", "content")
			text := tmpFs.File(tmpDir+pathSeparator+"text.txt", "text text text")

			watcher, e := watch.CreateWatch(1).
				WithHandlers(expectChangesOf(text.Name(), done), ignoreChangesOf(config.Name())).
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

		It("should not react on file change in excluded directory", func() {
			// given
			done := make(chan struct{})
			skipTmpDir := tmpFs.Dir("skip_watch")
			config := tmpFs.File(skipTmpDir+pathSeparator+"config.yaml", "content")
			watchTmpDir := tmpFs.Dir("watch")
			code := tmpFs.File(watchTmpDir+pathSeparator+"main.go", "package main")

			watcher, e := watch.CreateWatch(1).
				WithHandlers(ignoreChangesOf(config.Name()), expectChangesOf(code.Name(), done)).
				Excluding("skip_watch/**").
				OnPaths(filepath.Dir(skipTmpDir), filepath.Dir(watchTmpDir))
			Expect(e).ToNot(HaveOccurred())

			defer watcher.Close()

			// when
			watcher.Start()
			_, _ = config.WriteString(" should not be watched")
			_, _ = code.WriteString("\n // Bla!")

			// then
			Eventually(done).Should(BeClosed(), "Changes in the files has not been recognized!")
		})

		It("should not react on file change if it's git-ignored", func() {
			// given
			done := make(chan struct{})

			watchTmpDir := tmpFs.Dir("watch")
			config := tmpFs.File(watchTmpDir+pathSeparator+".idea"+pathSeparator+"config.toml", "content")
			test := tmpFs.File(watchTmpDir+pathSeparator+".idea"+pathSeparator+"test.yaml", "content")
			nestedFile := tmpFs.File(watchTmpDir+pathSeparator+"src"+pathSeparator+"main"+pathSeparator+"org"+pathSeparator+"acme"+pathSeparator+"Main.java", "package org.acme")
			gitIgnore := tmpFs.File(watchTmpDir+pathSeparator+".gitignore", `
*.yaml
**/src/main/**/*.java
.idea/
`)
			code := tmpFs.File(watchTmpDir+pathSeparator+"main.go", "package main")

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
				WithHandlers(ignoreChangesOf(config.Name(), test.Name(), nestedFile.Name()),
					expectChangesOf(code.Name(), done)).
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
})

// test doubles

func expectChangesOf(fileName string, done chan<- struct{}) watch.Handler {
	return func(events []fsnotify.Event) error {
		defer GinkgoRecover()
		Expect(events[0].Name).To(Equal(fileName))
		close(done)

		return nil
	}
}

func ignoreChangesOf(fileNames ...string) watch.Handler {
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
