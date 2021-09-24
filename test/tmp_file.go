package test

//  Based on https://github.com/Flaque/filet
//  We might want to switch to upstream when https://github.com/Flaque/filet/pull/2 gets merged

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/afero"
)

// TestReporter can be used to report test failures. It is satisfied by the standard library's *testing.T.
type TestReporter interface { //nolint:golint //reason test code
	Errorf(format string, args ...interface{})
}

type TmpFileSystem struct {
	// resources to keep track for cleanup
	resources []string
	fs        afero.Fs
	t         TestReporter
}

func NewTmpFileSystem(t TestReporter) TmpFileSystem {
	return TmpFileSystem{t: t, fs: afero.NewOsFs(), resources: []string{}}
}

// Dir creates a temporary directory under os.TempDir() with a following pattern:
// os.TempDir()/[random-alphanumeric]/dir, where dir is a passed parameter which can be a relative path
// When dir is an absolute path and error is reported.
func (tmp *TmpFileSystem) Dir(dir string) string {
	fullPath := dir
	if !path.IsAbs(dir) {
		// Removes trailing slash which is returned by MacOS https://github.com/golang/go/issues/21318
		// Otherwise ending up with `//` vs `/` in mac vs linux resulting in failures
		tmpDir := filepath.Clean(os.TempDir())
		fullPath = fmt.Sprintf("%s/%s/%s", tmpDir, randomAlphaNumeric(), dir)
	}

	if err := tmp.fs.MkdirAll(fullPath, os.ModePerm); err != nil {
		tmp.t.Errorf("Failed to create the directory: %s. Reason: %s", dir, err)

		return ""
	}

	tmp.resources = append(tmp.resources, fullPath)

	return fullPath
}

// File creates a specified file to use when testing
// if filePath is a full path it will just be created and cleaned up afterwards
// otherwise the file will be places under some random alphanumeric folder under temp directory.
func (tmp *TmpFileSystem) File(filePath, content string) afero.File {
	fullPath := filePath
	if !path.IsAbs(filePath) {
		// Removes trailing slash which is returned by MacOS https://github.com/golang/go/issues/21318
		// Otherwise ending up with `//` vs `/` in mac vs linux resulting in failures
		tmpDir := filepath.Clean(os.TempDir())
		fullPath = fmt.Sprintf("%s/%s/%s", tmpDir, randomAlphaNumeric(), filePath)
	}

	if err := tmp.fs.MkdirAll(path.Dir(fullPath), os.ModePerm); err != nil {
		tmp.t.Errorf("Failed to create the file: %s. Reason: %s", fullPath, err)

		return nil
	}

	file, err := tmp.fs.OpenFile(fullPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		tmp.t.Errorf("Failed to create the file: %s. Reason: %s", fullPath, err)

		return nil
	}

	if _, err := file.WriteString(content); err != nil {
		tmp.t.Errorf("Failed writing to a file")

		return nil
	}
	tmp.resources = append(tmp.resources, file.Name())

	return file
}

// Cleanup removes all files in our test registry and calls `t.Errorf` if something goes wrong.
func (tmp *TmpFileSystem) Cleanup() {
	for _, filePath := range tmp.resources {
		if err := tmp.fs.RemoveAll(filePath); err != nil {
			tmp.t.Errorf(tmp.fs.Name(), err)
		}
	}

	tmp.resources = make([]string, 0)
}

func randomAlphaNumeric() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	s := hex.EncodeToString(b)

	return s
}
