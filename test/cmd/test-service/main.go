package main

import (
	"flag"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"emperror.dev/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	"github.com/maistra/istio-workspace/pkg/log"
)

const (
	// EnvHTTPAddr name of Env variable that sets the listening address.
	EnvHTTPAddr = "HTTP_ADDR"

	// EnvGRPCAddr name of Env variable that sets the listening address.
	EnvGRPCAddr = "GRPC_ADDR"

	// EnvServiceName name of Env variable that sets the Service name used in call stack.
	EnvServiceName = "SERVICE_NAME"

	// EnvServiceCall name of Env variable that holds a colon-separated list of URLs to call.
	EnvServiceCall = "SERVICE_CALL"
)

var rootDir = "test/cmd/test-service/assets/" //nolint:varcheck,deadcode,unused //reason This is required to use the dev mode for assets (reading from fs)

// Config describes the basic Name and who to call next for a given HandlerFunc.
type Config struct {
	Name string
	Call []*url.URL
}

func main() {
	log.SetLogger(log.CreateOperatorAwareLogger("test").WithValues("type", "test-service"))

	c := Config{}
	if v, b := os.LookupEnv(EnvServiceName); b {
		c.Name = v
	}

	logger := log.Log.WithName("service").WithValues("name", c.Name)

	if v, b := os.LookupEnv(EnvServiceCall); b {
		u, err := parseURL(v)
		if err != nil {
			logger.Error(err, "Couldn't parse config")
			os.Exit(-1)
		}
		c.Call = u
	}

	serviceName := flag.String("serviceName", c.Name, "The service name")
	flag.Parse()

	if serviceName != nil {
		c.Name = *serviceName
	}

	httpAdr := "127.0.0.1:8080"
	if v, b := os.LookupEnv(EnvHTTPAddr); b {
		httpAdr = v
	}
	grpcAdr := "127.0.0.1:8081"
	if v, b := os.LookupEnv(EnvGRPCAddr); b {
		grpcAdr = v
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(http.StatusOK)
	})
	http.HandleFunc("/", NewBasic(c, MultiplexRequestInvoker, logger))
	go func() {
		err := http.ListenAndServe(httpAdr, nil)
		if err != nil {
			logger.Error(err, "failed initializing")
		}
		logger.Info("Started serving basic test service")
	}()

	lis, err := net.Listen("tcp", grpcAdr)
	if err != nil {
		logger.Error(err, "failed to listen")
	}
	s := grpc.NewServer()
	RegisterCallerServer(s, &server{Config: c, Invoker: MultiplexRequestInvoker, Log: logger})
	if err := s.Serve(lis); err != nil {
		logger.Error(err, "failed to serve")
	}
}

func parseURL(value string) ([]*url.URL, error) {
	vs := strings.Split(value, ",")
	urls := make([]*url.URL, 0, len(vs))
	for _, v := range vs {
		u, err := url.Parse(v)
		if err != nil {
			return nil, errors.Wrap(err, "failed parsing URL "+value)
		}
		urls = append(urls, u)
	}

	return urls, nil
}
