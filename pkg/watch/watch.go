package watch

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"emperror.dev/errors"
	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/logr"
	ignore "github.com/sabhiram/go-gitignore"

	"github.com/maistra/istio-workspace/pkg/log"
)

var logger = func() logr.Logger {
	return log.Log.WithValues("type", "watch")
}

// Handler allows to define how to react on file changes event.
type Handler func(events []fsnotify.Event) error

// Watch represents single file system watch and delegates change events to defined handler.
type Watch struct {
	watcher   *fsnotify.Watcher
	handlers  []Handler
	basePaths []string
	ignores   []ignore.GitIgnore
	interval  time.Duration
	done      chan struct{}
}

// Start observes on file change events and dispatches them to defined handler in batches every
// given interval.
func (w *Watch) Start() {
	// Dispatch fsnotify events
	go func() {
		tick := time.NewTicker(w.interval)
		events := make(map[string]fsnotify.Event)
	OutOfFor:
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if w.Excluded(event.Name) {
					logger().V(1).Info("file excluded. skipping change handling", "file", event.Name)

					continue
				}
				logger().V(1).Info("file changed", "file", event.Name, "op", event.Op.String())
				events[event.Name] = event
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				logger().Error(err, "failed while watching")
			case <-tick.C:
				if len(events) == 0 {
					continue
				}
				logger().V(1).Info("firing change event")
				changes := extractEvents(events)
				for _, handler := range w.handlers {
					if err := handler(changes); err != nil {
						logger().Error(err, "failed to handle file change!")
					}
				}
				events = make(map[string]fsnotify.Event)
			case <-w.done:
				break OutOfFor
			}
		}
		close(w.done)
	}()
}

// Excluded checks whether a path is excluded from watch by first inspecting .ignores
// and user-defined exclusions.
func (w *Watch) Excluded(path string) bool {
	for _, ignoreRule := range w.ignores {
		if ignoreRule.MatchesPath(path) {
			return true
		}
	}

	return false
}

// Close attempts to close underlying fsnotify.Watcher.
// In case of failure it logs the error.
func (w *Watch) Close() {
	w.done <- struct{}{}
	if e := w.watcher.Close(); e != nil {
		logger().Error(e, "failed closing watch")
	}
}

// addPath adds single path (non-recursive) to be watch.
func (w *Watch) addPath(filePath string) error {
	w.basePaths = append(w.basePaths, filePath)

	return errors.WrapIfWithDetails(w.watcher.Add(filePath), "failed adding path", "path", filePath)
}

// addRecursiveWatch handles adding watches recursively for the path provided
// and its subdirectories. If a non-directory is specified, this call is a no-op.
//
// Based on https://github.com/openshift/origin/blob/85eb37b3/pkg/util/fsnotification/fsnotification.go.
func (w *Watch) addRecursiveWatch(filePath string) error {
	file, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return errors.WithDetails(err, "error introspecting filePath", "path", filePath)
	}

	if !file.IsDir() {
		return nil
	}

	folders, err := allSubFoldersOf(filePath)
	if err != nil {
		return err
	}

	for _, v := range folders {
		logger().V(1).Info(fmt.Sprintf("adding watch on filePath %s", v))
		err = w.addPath(v)
		if err != nil {
			// "no space left on device" issues are usually resolved via
			// $ sudo sysctl fs.inotify.max_user_watches=65536
			return errors.WithDetails(err, "error adding watcher for filePath", "path", v)
		}
	}

	return nil
}

func (w *Watch) addExclusions(exclusions []string) error {
	if len(exclusions) == 0 {
		return nil
	}
	ignores, e := ignore.CompileIgnoreLines(exclusions...)
	if e != nil {
		return errors.Wrapf(e, "failed adding exclusion list %v", exclusions)
	}

	w.ignores = append(w.ignores, *ignores)

	return nil
}

// addGitIgnore adds .gitignore rules to the watcher if the file exists in the given path.
func (w *Watch) addGitIgnore(path string) error {
	gitIgnorePath := filepath.Join(path, ".gitignore")
	file, err := os.Open(gitIgnorePath)
	if err == nil {
		err := file.Close()
		if err != nil {
			return errors.WrapWithDetails(err, "failed closing file", "path", gitIgnorePath)
		}
		gitIgnore, err := ignore.CompileIgnoreFile(gitIgnorePath)
		if err != nil {
			return errors.WrapWithDetails(err, "failed compiling ignore list from .gitignore", "path", gitIgnorePath)
		}
		w.ignores = append(w.ignores, *gitIgnore)
	}

	return nil
}

// allSubFoldersOf recursively retrieves all subfolders of the specified path.
func allSubFoldersOf(filePath string) (paths []string, err error) {
	err = filepath.Walk(filePath, func(newPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			paths = append(paths, newPath)
		}

		return nil
	})

	return
}

// extractEvents takes a map and returns slice of all values.
func extractEvents(events map[string]fsnotify.Event) []fsnotify.Event {
	changes := make([]fsnotify.Event, len(events))
	i := 0
	for _, event := range events {
		changes[i] = event
		i++
	}

	return changes
}
