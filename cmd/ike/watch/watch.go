package watch

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("watch")

// Handler allows to define how to react on file changes even and if the watch should
// be terminated through closing done channel
type Handler func(event fsnotify.Event, done chan<- struct{}) error

type Watch struct {
	watcher    *fsnotify.Watcher
	handler    Handler
	exclusions FilePatterns
}

type Builder struct {
	w *Watch
}

func NewWatch() *Builder {
	return &Builder{w: &Watch{}}
}

func (wb *Builder) WithHandler(handler Handler) *Builder {
	wb.w.handler = handler
	return wb
}

func (wb *Builder) Excluding(exclusions ...string) *Builder {
	wb.w.exclusions = ParseFilePatterns(exclusions)
	return wb
}

func (wb *Builder) OnPaths(paths ...string) (watch *Watch, err error) {

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error(err, "failed creating fs watch")
		return nil, err
	}

	wb.w.watcher = fsWatcher

	for _, p := range paths {
		if e := wb.w.AddRecursiveWatch(p); e != nil {
			return nil, e
		}
	}

	return wb.w, nil
}

// AddPath adds single path (non-recursive) to be watch
func (w *Watch) AddPath(filePath string) error {
	return w.watcher.Add(filePath)
}

// AddRecursiveWatch handles adding watches recursively for the path provided
// and its subdirectories. If a non-directory is specified, this call is a no-op.
// Based on https://github.com/openshift/origin/blob/85eb37b34f0657631592356d020cef5a58470f8e/pkg/util/fsnotification/fsnotification.go
func (w *Watch) AddRecursiveWatch(filePath string) error {
	file, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error introspecting filePath %s: %v", filePath, err)
	}
	if !file.IsDir() {
		return nil
	}

	folders, err := getSubFolders(filePath)
	if err != nil {
		return err
	}
	for _, v := range folders {
		log.Info("adding watch on filePath %s", v)
		err = w.AddPath(v)
		if err != nil {
			// "no space left on device" issues are usually resolved via
			// $ sudo sysctl fs.inotify.max_user_watches=65536
			return fmt.Errorf("error adding watcher for filePath %s: %v", v, err)
		}
	}
	return nil
}

// getSubFolders recursively retrieves all subfolders of the specified path.
func getSubFolders(filePath string) (paths []string, err error) {
	err = filepath.Walk(filePath, func(newPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			paths = append(paths, newPath)
		}
		return nil
	})
	return paths, err
}

func (w *Watch) Close() {
	if e := w.watcher.Close(); e != nil {
		log.Error(e, "failed closing watch")
	}
}

func (w *Watch) Watch() <-chan struct{} {

	done := make(chan struct{})

	// Dispatch fsnotify events
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if w.exclusions.Matches(event.Name) {
					continue
				}

				if err := w.handler(event, done); err != nil {
					log.Error(err, "oups!")
				}
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Error(err, "failed while watching")
				close(done)
			}
		}
	}()

	// when done close watcher
	go func() {
		<-done
		w.Close()
	}()

	return done
}
