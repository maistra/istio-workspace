package watch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/maistra/istio-workspace/pkg/log"

	"github.com/fsnotify/fsnotify"
	ignore "github.com/sabhiram/go-gitignore"
)

var logger = log.CreateOperatorAwareLogger("watch")

// Handler allows to define how to react on file changes event.
type Handler func(events []fsnotify.Event) error

// Watch represents single file system watch and delegates change events to defined handler.
type Watch struct {
	watcher    *fsnotify.Watcher
	handlers   []Handler
	basePaths  []string
	gitignores []ignore.GitIgnore
	interval   time.Duration
	done       chan struct{}
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
					logger.V(1).Info("file excluded. skipping change handling", "file", event.Name)
					continue
				}
				logger.V(1).Info("file changed", "file", event.Name, "op", event.Op.String())
				events[event.Name] = event
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				logger.Error(err, "failed while watching")
			case <-tick.C:
				if len(events) == 0 {
					continue
				}
				logger.V(1).Info("firing change event")
				changes := extractEvents(events)
				for _, handler := range w.handlers {
					if err := handler(changes); err != nil {
						logger.Error(err, "failed to handle file change!")
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

// Excluded checks whether a path is excluded from watch by first inspecting .gitignores
// and user-defined exclusions.
func (w *Watch) Excluded(path string) bool {
	reducedPath := path
	for _, basePath := range w.basePaths {
		if strings.HasPrefix(path, basePath) {
			reducedPath = strings.TrimPrefix(path, basePath)
			break
		}
	}
	for _, gitIgnore := range w.gitignores {
		if gitIgnore.MatchesPath(reducedPath) {
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
		logger.Error(e, "failed closing watch")
	}
}

// addPath adds single path (non-recursive) to be watch.
func (w *Watch) addPath(filePath string) error {
	w.basePaths = append(w.basePaths, filePath)
	return w.watcher.Add(filePath)
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
		return fmt.Errorf("error introspecting filePath %s: %v", filePath, err)
	}

	if !file.IsDir() {
		return nil
	}

	folders, err := allSubFoldersOf(filePath)
	if err != nil {
		return err
	}

	for _, v := range folders {
		logger.V(1).Info(fmt.Sprintf("adding watch on filePath %s", v))
		err = w.addPath(v)
		if err != nil {
			// "no space left on device" issues are usually resolved via
			// $ sudo sysctl fs.inotify.max_user_watches=65536
			return fmt.Errorf("error adding watcher for filePath %s: %v", v, err)
		}
	}
	return nil
}

func (w *Watch) addExclusions(exclusions []string) error {
	if len(exclusions) == 0 {
		return nil
	}
	gitIgnore, e := ignore.CompileIgnoreLines(exclusions...)
	if e != nil {
		return e
	}
	w.gitignores = append(w.gitignores, *gitIgnore)
	return nil
}

// addGitIgnore adds .gitignore rules to the watcher if the file exists in the given path.
func (w *Watch) addGitIgnore(path string) error {
	gitIgnorePath := path + string(os.PathSeparator) + ".gitignore"
	file, err := os.Open(gitIgnorePath)
	if err == nil {
		err := file.Close()
		if err != nil {
			return err
		}
		gitIgnore, err := ignore.CompileIgnoreFile(gitIgnorePath)
		if err != nil {
			return err
		}
		w.gitignores = append(w.gitignores, *gitIgnore)
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
