package e2e

import (
	"io/ioutil"
	"net/http"
)

// GetBody calls GET on a given URL and returns its body or error in case there's one
func GetBody(rawURL string, cookies ...*http.Cookie) (string, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, err := http.DefaultClient.Do(req) //nolint:gosec //reason test code
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
	if v, f := headers["Host"]; f {
		req.Host = v
	}
	for k, v := range headers {
		req.Header[k] = []string{v}
	}
	resp, err := http.DefaultClient.Do(req) //nolint:gosec //reason cmd required by cobra
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	return string(content), nil
}
