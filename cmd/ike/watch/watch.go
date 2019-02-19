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

// Watch represents single file system watch and delegates change events to defined handler
type Watch struct {
	watcher    *fsnotify.Watcher
	handler    Handler
	exclusions FilePatterns
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
		log.Info(fmt.Sprintf("adding watch on filePath %s", v))
		err = w.AddPath(v)
		if err != nil {
			// "no space left on device" issues are usually resolved via
			// $ sudo sysctl fs.inotify.max_user_watches=65536
			return fmt.Errorf("error adding watcher for filePath %s: %v", v, err)
		}
	}
	return nil
}

// Close attempts to close underlying fsnotify.Watcher.
// In case of failure it logs the error
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
					log.Info("file excluded. skipping change handling", "file", event.Name)
					continue
				}

				if err := w.handler(event, done); err != nil {
					log.Error(err, "failed to handle file change!")
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
