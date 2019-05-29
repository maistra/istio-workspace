package watch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/denormal/go-gitignore"
	"github.com/fsnotify/fsnotify"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("watch")

// Handler allows to define how to react on file changes event
type Handler func(events []fsnotify.Event) error

// Watch represents single file system watch and delegates change events to defined handler
type Watch struct {
	watcher    *fsnotify.Watcher
	handlers   []Handler
	gitignores []gitignore.GitIgnore
	interval   time.Duration
	done       chan struct{}
}

// Start observes on file change events and dispatches them to defined handler in batches every
// given interval
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
					log.V(10).Info("file excluded. skipping change handling", "file", event.Name)
					continue
				}
				log.Info("file changed", "file", event.Name, "op", event.Op.String())
				events[event.Name] = event
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Error(err, "failed while watching")
			case <-tick.C:
				if len(events) == 0 {
					continue
				}
				log.Info("firing change event")
				changes := extractValues(events)
				for _, handler := range w.handlers {
					if err := handler(changes); err != nil {
						log.Error(err, "failed to handle file change!")
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

// Excluded checks whether a path is excluded from watch by first inspecting .gitignores and only then user-defined
// exclusions
func (w *Watch) Excluded(path string) bool {
	for _, ignore := range w.gitignores {
		if ignore.Ignore(path) {
			return true
		}
	}
	return false
}

// Close attempts to close underlying fsnotify.Watcher.
// In case of failure it logs the error
func (w *Watch) Close() {
	w.done <- struct{}{}
	if e := w.watcher.Close(); e != nil {
		log.Error(e, "failed closing watch")
	}
}

// addPath adds single path (non-recursive) to be watch
func (w *Watch) addPath(filePath string) error {
	return w.watcher.Add(filePath)
}

// addRecursiveWatch handles adding watches recursively for the path provided
// and its subdirectories. If a non-directory is specified, this call is a no-op.
// Based on https://github.com/openshift/origin/blob/85eb37b34f0657631592356d020cef5a58470f8e/pkg/util/fsnotification/fsnotification.go
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

	folders, err := getSubFolders(filePath)
	if err != nil {
		return err
	}

	for _, v := range folders {
		log.V(3).Info(fmt.Sprintf("adding watch on filePath %s", v))
		err = w.addPath(v)
		if err != nil {
			// "no space left on device" issues are usually resolved via
			// $ sudo sysctl fs.inotify.max_user_watches=65536
			return fmt.Errorf("error adding watcher for filePath %s: %v", v, err)
		}
	}
	return nil
}

func (w *Watch) addExclusions(base string, exclusions []string) {
	ignore := gitignore.New(strings.NewReader(strings.Join(exclusions, "\n")), base, nil)
	w.gitignores = append(w.gitignores, ignore)
}

// addGitIgnore adds .gitignore rules to the watcher if the file exists in the given path
func (w *Watch) addGitIgnore(path string) error {
	gitIgnorePath := path + string(os.PathSeparator) + ".gitignore"
	file, err := os.Open(gitIgnorePath)
	if err == nil {
		err := file.Close()
		if err != nil {
			return err
		}
		ignore, err := gitignore.NewFromFile(gitIgnorePath)
		if err != nil {
			return err
		}
		w.gitignores = append(w.gitignores, ignore)
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
	return
}

// extractValues takes a map and returns slice of all values
func extractValues(events map[string]fsnotify.Event) []fsnotify.Event {
	changes := make([]fsnotify.Event, len(events))
	i := 0
	for _, event := range events {
		changes[i] = event
		i++
	}
	return changes
}
