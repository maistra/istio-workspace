package infra

import (
	"fmt"
	"os"

	"github.com/maistra/istio-workspace/test/shell"

	"github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var appFs = afero.NewOsFs()

// CreateFile creates file under defined path with a given content
func CreateFile(filePath, content string) {
	file, err := appFs.Create(filePath)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	err = appFs.Chmod(filePath, os.ModePerm)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	_, err = file.WriteString(content)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() {
		err = file.Close()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	}()
}

// DeleteFile deletes file under defined path
func DeleteFile(filePath string) {
	err := appFs.Remove(filePath)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
}

func NewProjectCmd(name string) string {
	return fmt.Sprintf(`oc new-project %s --description "%s"`, name, "istio-workspace test project")
}

func DeleteProjectCmd(name string) string {
	return fmt.Sprintf(`oc delete project %s`, name)
}

// CreateNewApp creates new project with a given name, deploys simple datawire/hello-world app and exposes route to
// it service
func CreateNewApp(name string) {
	<-shell.Execute(NewProjectCmd(name)).Done()

	UpdateSecurityConstraintsFor(name)

	<-shell.ExecuteInDir(".",
		"oc", "new-app",
		"--docker-image", "datawire/hello-world",
		"--name", name,
		"--namespace", name,
		"--allow-missing-images",
	).Done()
	shell.ExecuteAll("oc expose svc/"+name, "oc status")
}

func DeployHelloWorldCmd(name, ns string) string {
	return "oc new-app --docker-image datawire/hello-world " +
		"--name " + name + " " +
		"--namespace " + ns + " " +
		"--allow-missing-images"
}
