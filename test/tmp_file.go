package test

//  Based on https://github.com/Flaque/filet
//  We might want to switch to upstream when https://github.com/Flaque/filet/pull/2 gets merged

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path"

	"github.com/spf13/afero"
)

// TestReporter can be used to report test failures. It is satisfied by the standard library's *testing.T.
type TestReporter interface { //nolint[:golint]
	Errorf(format string, args ...interface{})
}

// Files keeps track of files that we've used so we can clean up.
var Files []string
var appFs = afero.NewOsFs()

// TmpFile Creates a specified file for us to use when testing
func TmpFile(t TestReporter, fileName, content string) afero.File {
	filePath := fmt.Sprintf("%s/%s/%s", os.TempDir(), randomAlphaNumeric(), fileName)

	if err := appFs.MkdirAll(path.Dir(filePath), os.ModePerm); err != nil {
		t.Errorf("Failed to create the file: %s. Reason: %s", filePath, err)
		return nil
	}

	file, err := appFs.OpenFile(filePath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		t.Errorf("Failed to create the file: %s. Reason: %s", filePath, err)
		return nil
	}

	if _, err := file.WriteString(content); err != nil {
		t.Errorf("Failed writing to a file")
		return nil
	}
	Files = append(Files, file.Name())

	return file
}

// CleanUp removes all files in our test registry and calls `t.Errorf` if something goes wrong.
func CleanUp(t TestReporter) {
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
