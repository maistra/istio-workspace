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

// Files keeps track of files that we've used so we can clean up.
var Files []string

var appFs = afero.NewOsFs()

// TmpDir creates a temporary directory under os.TempDir() with a following pattern:
// os.TempDir()/[random-alphanumeric]/dir, where dir is a passed parameter which can be a relative path
// When dir is an absolute path and error is reported.
func TmpDir(t TestReporter, dir string) string {
	fullPath := dir
	if !path.IsAbs(dir) {
		// Removes trailing slash which is returned by MacOS https://github.com/golang/go/issues/21318
		// Otherwise ending up with `//` vs `/` in mac vs linux resulting in failures
		tmpDir := filepath.Clean(os.TempDir())
		fullPath = fmt.Sprintf("%s/%s/%s", tmpDir, randomAlphaNumeric(), dir)
	}

	if err := appFs.MkdirAll(fullPath, os.ModePerm); err != nil {
		t.Errorf("Failed to create the directory: %s. Reason: %s", dir, err)
		return ""
	}

	Files = append(Files, fullPath)

	return fullPath
}

// TmpFile creates a specified file for us to use when testing
// if filePath is a full path it will just be created and cleaned up afterwards
// otherwise the file will be places under some random alphanumeric folder under temp directory.
func TmpFile(t TestReporter, filePath, content string) afero.File {
	fullPath := filePath
	if !path.IsAbs(filePath) {
		// Removes trailing slash which is returned by MacOS https://github.com/golang/go/issues/21318
		// Otherwise ending up with `//` vs `/` in mac vs linux resulting in failures
		tmpDir := filepath.Clean(os.TempDir())
		fullPath = fmt.Sprintf("%s/%s/%s", tmpDir, randomAlphaNumeric(), filePath)
	}

	if err := appFs.MkdirAll(path.Dir(fullPath), os.ModePerm); err != nil {
		t.Errorf("Failed to create the file: %s. Reason: %s", fullPath, err)
		return nil
	}

	file, err := appFs.OpenFile(fullPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		t.Errorf("Failed to create the file: %s. Reason: %s", fullPath, err)
		return nil
	}

	if _, err := file.WriteString(content); err != nil {
		t.Errorf("Failed writing to a file")
		return nil
	}
	Files = append(Files, file.Name())

	return file
}

// CleanUpTmpFiles removes all files in our test registry and calls `t.Errorf` if something goes wrong.
func CleanUpTmpFiles(t TestReporter) {
	for _, filePath := range Files {
		if err := appFs.RemoveAll(filePath); err != nil {
			t.Errorf(appFs.Name(), err)
		}
	}

	Files = make([]string, 0)
}

func randomAlphaNumeric() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	s := hex.EncodeToString(b)
	return s
}
