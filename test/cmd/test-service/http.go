package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

// NewBasic constructs a new basic HandlerFunc that behaves as described by the provided Config.
func NewBasic(config Config, invoker RequestInvoker, log logr.Logger) http.HandlerFunc {
	return basic(config, invoker, log)
}

func basic(config Config, invoker RequestInvoker, log logr.Logger) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		logIncomingRequest(log, req)
		if strings.Contains(req.Header.Get("accept"), "text/html") {
			b, err := Asset("index.html")
			if err != nil {
				resp.WriteHeader(500)
				_, _ = resp.Write([]byte(err.Error()))
				return
			}
			resp.Header().Set("content-type", "text/html")
			resp.WriteHeader(200)
			_, _ = resp.Write(b)
			return
		}

		start := time.Now()
		callStack := CallStack{
			Caller:    config.Name,
			Protocol:  "http",
			Path:      req.URL.Path,
			StartTime: start.UnixNano(),
			Color:     "#FFF",
		}

		for _, u := range config.Call {
			func() {
				called := invoker(log, u, getHeaders(req, propagationHeaders...))
				if called != nil {
					callStack.Called = append(callStack.Called, called)
				}
			}()
		}

		end := time.Now()
		callStack.EndTime = end.UnixNano()

		resp.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(resp)
		enc.SetIndent("", "  ")
		var err = enc.Encode(&callStack)
		if err != nil {
			fmt.Println("Failed to encode", err)
			return
		}
	}
}

func httpRequestInvoker(log logr.Logger, target *url.URL, headers map[string]string) *CallStack {
	request, err := http.NewRequest("GET", target.String(), nil)
	if err != nil {
		log.Error(err, "Failed to create request", "target", target)
		return nil
	}
	copyHeaders(request, headers)
	logOutgoingRequest(log, request)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Error(err, "Failed to call", "target", target)
		return nil
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	child := CallStack{}
	err = dec.Decode(&child)
	if err != nil {
		log.Error(err, "Failed to decode", "target", target)
		return nil
	}
	return &child
}

func getHeaders(source *http.Request, headers ...string) map[string]string {
	m := map[string]string{}
	for _, header := range headers {
		if value := source.Header.Get(header); value != "" {
			m[header] = value
		}
	}
	return m
}

func copyHeaders(target *http.Request, headers map[string]string) {
	for key, value := range headers {
		target.Header.Set(key, value)
	}
}

func logIncomingRequest(log logr.Logger, req *http.Request) {
	log.Info("received", "protocol", "http", "path", req.URL.Path, "headers", req.Header)
}

func logOutgoingRequest(log logr.Logger, req *http.Request) {
	log.Info("sent", "protocol", "http", "target", req.URL.String(), "path", req.URL.Path, "headers", req.Header)
}
