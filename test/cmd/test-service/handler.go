package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Config describes the basic Name and who to call next for a given HandlerFunc
type Config struct {
	Name string
	Call []*url.URL
}

// CallStack holds the complete transaction stack
type CallStack struct {
	Caller    string      `json:"caller"`
	Path      string      `json:"path"`
	StartTime time.Time   `json:"startTime"`
	EndTime   time.Time   `json:"endTime"`
	Called    []CallStack `json:"called,omitempty"`
}

// NewBasic constructs a new basic HandlerFunc that behaves as described by the provided Config
func NewBasic(config Config) http.HandlerFunc {
	return basic(config)
}

func basic(config Config) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		start := time.Now()
		callStack := CallStack{
			Caller:    config.Name,
			Path:      req.URL.Path,
			StartTime: start,
		}

		for _, url := range config.Call {
			func() {
				request, err := http.NewRequest("GET", url.String(), nil)
				if err != nil {
					fmt.Println("Failed to create request", url, err)
					return
				}
				copyHeaders(
					request,
					req,
					"x-request-id", "x-b3-traceid", "x-b3-spanid",
					"x-b3-parentspanid", "x-b3-sampled", "x-b3-flags",
					"x-ot-span-context")

				// custom header used by test suite
				copyHeaders(request, req, "x-test-suite")
				resp, err := http.DefaultClient.Do(request)
				if resp != nil {
					defer resp.Body.Close()
				}
				if err != nil {
					fmt.Println("Failed to call", url, err)
					return
				}
				dec := json.NewDecoder(resp.Body)
				child := CallStack{}
				err = dec.Decode(&child)
				if err != nil {
					fmt.Println("Failed to decode", url, err)
					return
				}
				callStack.Called = append(callStack.Called, child)
			}()
		}

		end := time.Now()
		callStack.EndTime = end

		resp.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(resp)
		enc.SetIndent("", "  ")
		err := enc.Encode(callStack)
		if err != nil {
			fmt.Println("Failed to encode", err)
			return
		}
	}
}

func copyHeaders(target, source *http.Request, headers ...string) {
	for _, header := range headers {
		if value := source.Header.Get(header); value != "" {
			target.Header.Set(header, value)
		}
	}
}
