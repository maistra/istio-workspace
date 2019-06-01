package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// EnvHTTPAddr name of Env variable that sets the listning address
	EnvHTTPAddr = "HTTP_ADDR"

	// EnvServiceName name of Env variable that sets the Service name used in call stack
	EnvServiceName = "SERVICE_NAME"

	// EnvServiceCall name of Env variable that holds a colon separated list of URLs to call
	EnvServiceCall = "SERVICE_CALL"
)

func main() {
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

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusOK)
	})
	http.HandleFunc("/", NewBasic(c))
	err := http.ListenAndServe(adr, nil)
	if err != nil {
		fmt.Println(err)
	}
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
