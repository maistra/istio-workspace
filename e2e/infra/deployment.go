package infra

import (
	"os"

	"github.com/spf13/afero"
)

var appFs = afero.NewOsFs()

// CreateFile creates file under defined path with a given content.
func CreateFile(filePath, content string) {
	file, err := appFs.Create(filePath)

	defer func() {
		if err = file.Close(); err != nil {
			panic(err)
		}
	}()

	if err != nil {
		panic(err)
	}

	if err = appFs.Chmod(filePath, os.ModePerm); err != nil {
		panic(err)
	}

	if _, err = file.WriteString(content); err != nil {
		panic(err)
	}
}

// DeleteFile deletes file under defined path.
func DeleteFile(filePath string) {
	if err := appFs.Remove(filePath); err != nil {
		panic(err)
	}
}

func NewProjectCmd(name string) string {
	return "kubectl create namespace " + name
}

func DeleteProjectCmd(name string) string {
	return "kubectl delete namespace " + name + " --wait=false"
}

func DeployNoopLoopCmd(name, ns string) []string {
	return []string{
		"kubectl run " + name +
			" --image=crccheck/hello-world" +
			" --port 8000" +
			" --expose" +
			" --namespace " + ns,
		"kubectl create deployment " + name +
			" --image=crccheck/hello-world" +
			" --namespace " + ns,
	}
}
