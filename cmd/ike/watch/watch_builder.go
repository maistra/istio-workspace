package watch

import (
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Builder is a struct which allows to use fluent API to create underlying instance of Watch
type Builder struct {
	w *Watch
}

// CreateWatch creates instance of Builder providing fluent functions to customize watch
// 		interval defines how frequently (in ms) file change events should be processed
func CreateWatch(intervalMs int64) *Builder {
	return &Builder{w: &Watch{
		interval: time.Duration(intervalMs) * time.Millisecond,
		done:     make(chan struct{}, 1),
	}}
}

// WithHandlers allows to hook instance of Handler reacting on file change events
func (wb *Builder) WithHandlers(handlers ...Handler) *Builder {
	wb.w.handlers = handlers
	return wb
}

// Excluding allows to define exclusion patterns (as glob expressions)
func (wb *Builder) Excluding(exclusions ...string) *Builder {
	wb.w.exclusions = ParseFilePatterns(exclusions)
	return wb
}

// OnPaths defines paths to be watched.
// If path is a directory it will recursively watch all files and subdirectories.
// If path is a file only this path is watched
func (wb *Builder) OnPaths(paths ...string) (watch *Watch, err error) {

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error(err, "failed creating fs watch")
		return nil, err
	}

	wb.w.watcher = fsWatcher

	for _, p := range paths {
		dir, err := os.Stat(p)

		if err != nil {
			return nil, err
		}

		if !dir.IsDir() {
			if e := wb.w.addPath(p); e != nil {
				return nil, e
			}
		} else {
			if e := wb.w.addRecursiveWatch(p); e != nil {
				return nil, e
			}
		}

	}

	return wb.w, nil
}
