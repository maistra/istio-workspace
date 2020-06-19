package e2e

import (
	"io/ioutil"
	"net/http"
)

// GetBodyWithHeaders calls GET on a given URL with a specific set request headers and returns its body or error in case there's one.
func GetBodyWithHeaders(rawURL string, headers map[string]string) (string, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	if v, f := headers["Host"]; f {
		req.Host = v
	}
	for k, v := range headers {
		req.Header[k] = []string{v}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	return string(content), nil
}
