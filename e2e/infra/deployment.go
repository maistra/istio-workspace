package infra

import (
	"os"

	"github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var appFs = afero.NewOsFs()

// CreateFile creates file under defined path with a given content.
func CreateFile(filePath, content string) {
	file, err := appFs.Create(filePath)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	err = appFs.Chmod(filePath, os.ModePerm)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	_, err = file.WriteString(content)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	err = file.Close()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
}

// DeleteFile deletes file under defined path.
func DeleteFile(filePath string) {
	err := appFs.Remove(filePath)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
}

func NewProjectCmd(name string) string {
	return "kubectl create namespace " + name
}

func DeleteProjectCmd(name string) string {
	return "kubectl delete namespace " + name + " --wait=false"
}

func DeployHelloWorldCmd(name, ns string) []string {
	return []string{
		"kubectl run " + name +
			" --image=datawire/hello-world" +
			" --port 8000" +
			" --expose" +
			" --namespace " + ns,
		"kubectl create deployment " + name +
			" --image=datawire/hello-world" +
			" --namespace " + ns,
	}
}
