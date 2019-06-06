package infra

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var appFs = afero.NewOsFs()

// ModifyServerCodeIn changes the code base of a simple python-based web server and puts it in the defined directory
func ModifyServerCodeIn(tmpDir string) {
	CreateFile(tmpDir+"/"+"server.py", ModifiedServerPy)
}

// OriginalServerCodeIn puts the original code base of a simple python-based web server in the defined directory
func OriginalServerCodeIn(tmpDir string) {
	CreateFile(tmpDir+"/"+"server.py", OrigServerPy)
}

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

// GetBody calls GET on a given URL and returns its body or error in case there's one
func GetBody(rawURL string, cookies ...*http.Cookie) (string, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, err := http.DefaultClient.Do(req) //nolint[:gosec]
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	return string(content), nil
}

// GetBodyWithHeaders calls GET on a given URL with a specific set request headers and returns its body or error in case there's one
func GetBodyWithHeaders(rawURL string, headers map[string]string) (string, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	for k, v := range headers {
		req.Header[k] = []string{v}
	}
	resp, err := http.DefaultClient.Do(req) //nolint[:gosec]
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	return string(content), nil
}
