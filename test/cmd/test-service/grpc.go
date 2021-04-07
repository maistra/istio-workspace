package main

import (
	"context"
	"net/url"
	"time"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type server struct {
	Config  Config
	Invoker RequestInvoker
	Log     logr.Logger
}

func (s server) Call(ctx context.Context, callee *Callee) (*CallStack, error) {
	headers := getMetadata(ctx, propagationHeaders...)
	s.Log.Info("received", "protocol", "grpc", "headers", headers)

	start := time.Now()
	callStack := CallStack{
		Caller:    s.Config.Name,
		Protocol:  "grpc",
		Path:      "",
		StartTime: start.UnixNano(),
		Color:     "#FFF",
	}

	for _, callee := range s.Config.Call {
		callStack.Called = append(callStack.Called, s.Invoker(s.Log, callee, headers))
	}

	end := time.Now()
	callStack.EndTime = end.UnixNano()

	return &callStack, nil
}

func gRPCRequestInvoker(log logr.Logger, target *url.URL, headers map[string]string) *CallStack {
	log.Info("sent", "protocol", "grpc", "target", target.String(), "headers", headers)
	conn, err := grpc.Dial(target.String(), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Error(err, "Failed to connect", "target", target)

		return nil
	}
	defer conn.Close()
	c := NewCallerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, mapToArray(headers)...)
	r, err := c.Call(ctx, &Callee{})
	if err != nil {
		log.Error(err, "Failed to call service", "target", target)
	}

	return r
}

func mapToArray(m map[string]string) []string {
	s := make([]string, 0, len(m))
	for k, v := range m {
		s = append(s, k, v)
	}

	return s
}

func getMetadata(ctx context.Context, headers ...string) map[string]string {
	m := map[string]string{}
	if meta, ok := metadata.FromIncomingContext(ctx); ok {
		for _, header := range headers {
			if value := meta.Get(header); len(value) != 0 {
				m[header] = value[0]
			}
		}
	}

	return m
}
