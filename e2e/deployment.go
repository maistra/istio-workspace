package e2e

import (
	"github.com/spf13/afero"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"

	. "github.com/onsi/gomega"
)

func DownloadInto(dir string, rawDownloadUrl string) string {
	content, err := GetBody(rawDownloadUrl)
	Expect(err).ToNot(HaveOccurred())

	downloadUrl, err := url.Parse(rawDownloadUrl)
	Expect(err).ToNot(HaveOccurred())

	filePath := dir + "/" + path.Base(downloadUrl.Path)
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
	Expect(err).NotTo(HaveOccurred())
	err = appFs.Chmod(filePath, os.ModePerm)
	Expect(err).ToNot(HaveOccurred())
	_, err = file.WriteString(content)
	Expect(err).ToNot(HaveOccurred())
	defer func() {
		err = file.Close()
		Expect(err).ToNot(HaveOccurred())
	}()
}

func GetBody(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	return string(content), nil
}
