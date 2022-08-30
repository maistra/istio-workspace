package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

// Simple http test service which returns HOSTNAME, which, when ran in k8s will be the pod name.
func main() {
	port := flag.String("port", "8181", "The address this service is available on")
	flag.Parse()

	log.Println("Starting new service on", *port)

	if err := http.ListenAndServe(":"+*port, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		statusCode := http.StatusOK
		response, found := os.LookupEnv("KUBERNETES_SERVICE_PORT")

		if !found {
			response = "not in k8s"
			statusCode = http.StatusNotFound
		} else {
			response = "running in k8s: " + response
		}

		writer.WriteHeader(statusCode)
		_, _ = writer.Write([]byte(response))
	})); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
