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
		podName, found := os.LookupEnv("HOSTNAME")

		if !found {
			podName = "not-running-in-k8s"
			statusCode = http.StatusNotFound
		}

		writer.WriteHeader(statusCode)
		_, _ = writer.Write([]byte(podName))
	})); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
