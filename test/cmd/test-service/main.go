package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	// EnvHTTPAddr name of Env variable that sets the listening address
	EnvHTTPAddr = "HTTP_ADDR"

	// EnvServiceName name of Env variable that sets the Service name used in call stack
	EnvServiceName = "SERVICE_NAME"

	// EnvServiceCall name of Env variable that holds a colon-separated list of URLs to call
	EnvServiceCall = "SERVICE_CALL"
)

var (
	rootDir = "test/cmd/test-service/assets/" //nolint[:deadcode]
)

func main() {
	logf.SetLogger(logf.ZapLogger(false))

	c := Config{}
	if v, b := os.LookupEnv(EnvServiceName); b {
		c.Name = v
	}
	if v, b := os.LookupEnv(EnvServiceCall); b {
		u, err := parseURL(v)
		if err != nil {
			fmt.Println("Couldn't parse config", err)
			os.Exit(-1)
		}
		c.Call = u
	}

	adr := ":8080"
	if v, b := os.LookupEnv(EnvHTTPAddr); b {
		adr = v
	}

	log := logf.Log.WithName("service").WithValues("name", c.Name)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusOK)
	})
	http.HandleFunc("/", NewBasic(c, log))
	err := http.ListenAndServe(adr, nil)
	if err != nil {
		log.Error(err, "failed initializing")
	}
	log.Info("Started serving basic test service")
}

func parseURL(value string) ([]*url.URL, error) {
	urls := []*url.URL{}
	vs := strings.Split(value, ",")
	for _, v := range vs {
		u, err := url.Parse(v)
		if err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}

	return urls, nil
}
