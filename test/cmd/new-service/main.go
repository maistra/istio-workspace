package main

import (
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
		resp, err := http.Get(*target)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(err.Error()))
			return
		}

		writer.WriteHeader(resp.StatusCode)
		body := resp.Body
		defer body.Close()
		content, _ := ioutil.ReadAll(body)
		writer.Write(content)
	})); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
