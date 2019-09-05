package version

import (
	"net/http"

	"github.com/google/go-github/github"
	"golang.org/x/net/context"
)

func LatestRelease() (string, error) {
	httpClient := http.Client{}
	defer httpClient.CloseIdleConnections()

	client := github.NewClient(&httpClient)
	latestRelease, _, e := client.Repositories.
		GetLatestRelease(context.Background(), "maistra", "istio-workspace")
	if e != nil {
		return "", e
	}
	return *latestRelease.Name, nil
}
