package verify

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"emperror.dev/errors"
	"github.com/schollz/progressbar/v3"
)

// GetBodyWithHeaders calls GET on a given URL with a specific set request headers
// and returns its body or error in case there's one.
func GetBodyWithHeaders(rawURL string, headers map[string]string) (string, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", rawURL, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed creating request")
	}
	if v, f := headers["Host"]; f {
		req.Host = v
	}
	for k, v := range headers {
		req.Header[k] = []string{v}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.WrapWithDetails(err, "failed executing HTTP call", "url", req.URL.String())
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)

	return string(content), nil
}

func call(routeURL string, headers map[string]string) func() (string, error) {
	fmt.Printf("Checking [%s] with headers [%s]\n", routeURL, headers)
	bar := progressbar.Default(-1)

	return func() (string, error) {
		if err := bar.Add(1); err != nil {
			return "", err
		}

		return GetBodyWithHeaders(routeURL, headers)
	}
}
