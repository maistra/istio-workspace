package test

import (
	"os"
	"strings"
)

type TmpPath struct {
	originalPath string
}

func NewTmpPath() *TmpPath {
	return &TmpPath{originalPath: os.Getenv("PATH")}
}

func (t *TmpPath) SetPath(paths ...string)  {
	_ = os.Setenv("PATH", strings.Join(paths, ":"))
}

func (t *TmpPath) Restore() {
	_ = os.Setenv("PATH", t.originalPath)
}
