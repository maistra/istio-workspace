FROM registry.access.redhat.com/ubi8/go-toolset:1.14.7 AS builder

USER 0

RUN mkdir -p /opt/app-root/src/go/github.com/maistra/istio-workspace/

WORKDIR /opt/app-root/src/go/github.com/maistra/istio-workspace/

COPY . .

ENV GOBIN=/opt/app-root/src/go/bin/

## GOROOT needs to be set explictly. see https://github.com/operator-framework/operator-sdk/issues/1854#issuecomment-525169969
RUN GOROOT=$(go env GOROOT) GOPATH=$(go env GOPATH) PATH="${PATH}:${GOPATH}/bin" make deps tools compile

## Controller image
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.3

COPY --from=builder /opt/app-root/src/go/github.com/maistra/istio-workspace/dist/ike /usr/local/bin/
