package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
)

// Simple web server to verify if we can reach services in the cluster when running it locally.
// Used for "ike develop new" scenario.
func main() {
	port := flag.String("port", ":8181", "The address this service is available on")
	target := flag.String("target", "http://reviews:9080", "The target service to be called")
	flag.Parse()

	log.Println("Starting new service on", *port)

	if err := http.ListenAndServe(*port, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, *target, nil)
		if err != nil {
			respondWithErr(writer, err)

			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			respondWithErr(writer, err)

			return
		}

		writer.WriteHeader(resp.StatusCode)
		body := resp.Body
		defer body.Close()
		content, _ := ioutil.ReadAll(body)
		_, _ = writer.Write(content)
	})); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func respondWithErr(writer http.ResponseWriter, e error) {
	writer.WriteHeader(http.StatusInternalServerError)
	_, _ = writer.Write([]byte(e.Error()))
}
