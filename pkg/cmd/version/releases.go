package version

import (
	"net/http"

	"github.com/maistra/istio-workspace/version"

	"github.com/google/go-github/github"
	"golang.org/x/net/context"
)

func LatestRelease() (string, error) {
	httpClient := http.Client{}
	defer httpClient.CloseIdleConnections()

	client := github.NewClient(&httpClient)
	latestRelease, _, err := client.Repositories.
		GetLatestRelease(context.Background(), "maistra", "istio-workspace")
	if err != nil {
		logger.Error(err, "unable to determine latest released version")
		return "", err
	}
	return *latestRelease.Name, nil
}

func IsLatestRelease(latestRelease string) bool {
	return latestRelease == version.Version
}
