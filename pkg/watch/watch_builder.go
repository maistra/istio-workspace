package watch

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"emperror.dev/errors"
	"github.com/fsnotify/fsnotify"
)

// Builder is a struct which allows to use fluent API to create underlying instance of Watch.
type Builder struct {
	w          *Watch
	exclusions []string
}

// CreateWatch creates instance of Builder providing fluent functions to customize watch
// 		interval defines how frequently (in ms) file change events should be processed.
func CreateWatch(intervalMs int64) *Builder {
	return &Builder{w: &Watch{
		interval: time.Duration(intervalMs) * time.Millisecond,
		done:     make(chan struct{}, 1),
	}}
}

// WithHandlers allows to register instances of Handler which will react on file change events.
func (wb *Builder) WithHandlers(handlers ...Handler) *Builder {
	wb.w.handlers = handlers

	return wb
}

// Excluding allows to define exclusion patterns (as glob expressions).
func (wb *Builder) Excluding(exclusions ...string) *Builder {
	wb.exclusions = exclusions

	return wb
}

// OnPaths defines paths to be watched.
// If path is a directory it will recursively watch all files and subdirectories.
// If path is a file only this path is watched.
func (wb *Builder) OnPaths(paths ...string) (watch *Watch, err error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrap(err, "failed creating fs watch")
	}

	wb.w.watcher = fsWatcher

	for _, path := range paths {
		dir, err := os.Stat(path)
		if err != nil {
			return nil, errors.WrapWithDetails(err, "failed checking path", "path", path)
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, errors.WrapWithDetails(err, "failed determining absolute path", "path", path)
		}
		if dir.IsDir() {
			if e := wb.w.addExclusions(wb.exclusions); e != nil {
				return nil, e
			}
			if e := wb.w.addGitIgnore(absPath); e != nil {
				return nil, e
			}
			if e := wb.w.addRecursiveWatch(absPath); e != nil {
				return nil, e
			}
		} else {
			e := wb.w.addPath(absPath)
			if e != nil {
				return nil, e
			}
		}
	}

	return wb.w, nil
}
