package infra

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/onsi/gomega"
	"github.com/spf13/afero"
)

func DownloadInto(dir, rawDownloadURL string) string {
	content, err := GetBody(rawDownloadURL)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	downloadURL, err := url.Parse(rawDownloadURL)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	filePath := dir + "/" + path.Base(downloadURL.Path)
	CreateFile(filePath, content)

	return filePath
}

var appFs = afero.NewOsFs()

func ModifyServerCodeIn(tmpDir string) {
	CreateFile(tmpDir+"/"+"server.py", ModifiedServerPy)
}

func OriginalServerCodeIn(tmpDir string) {
	CreateFile(tmpDir+"/"+"server.py", OrigServerPy)
}

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

func GetBody(rawURL string) (string, error) {
	resp, err := http.Get(rawURL) //nolint[:gosec]
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	return string(content), nil
}
