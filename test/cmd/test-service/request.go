package main

import (
	"net/url"
	"strings"

	"github.com/go-logr/logr"
)

// RequestInvoker calls a HTTP or gRPC URL.
type RequestInvoker func(log logr.Logger, target *url.URL, headers map[string]string) *CallStack

// MultiplexRequestInvoker switches Client impl based on URL schema. Supports http or grpc.
func MultiplexRequestInvoker(log logr.Logger, target *url.URL, headers map[string]string) *CallStack {
	if strings.Contains(target.Scheme, "http") {
		return httpRequestInvoker(log, target, headers)
	}
	return gRPCRequestInvoker(log, target, headers)
}
