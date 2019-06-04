package infra

import (
	"io/ioutil"
	"net/http"
	"net/url"
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

// PostBody issues a POST to the specified URL,
// with data's keys and values URL-encoded as the request body.
//
// Returns response's content, cookies or error if the POST failed
func PostBody(rawURL string, data url.Values, follow bool) (string, []*http.Cookie, error) {
	client := &http.Client{}
	if !follow {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	resp, err := client.PostForm(rawURL, data) //nolint[:gosec]
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	cookies := resp.Cookies()

	return string(content), cookies, nil
}
