package test

import (
	"os"
	"strings"
)

// TmpPath lets you overwrite $PATH environment variable for the duration of tests.
type TmpPath struct {
	originalPath string
}

// NewTmpPath creates new instance of TmpPath with stored original value assigned to $PATH environment variable.
func NewTmpPath() *TmpPath {
	return &TmpPath{originalPath: os.Getenv("PATH")}
}

// SetPath defines $PATH value.
func (t *TmpPath) SetPath(paths ...string) {
	_ = os.Setenv("PATH", strings.Join(paths, ":"))
}

// Restore restores $PATH to its original value from before creation of TmpPath instance.
func (t *TmpPath) Restore() {
	_ = os.Setenv("PATH", t.originalPath)
}
