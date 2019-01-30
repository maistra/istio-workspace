FROM centos:7

RUN adduser istio-workspace
USER istio-workspace

ADD dist/istio-workspace /usr/local/bin/istio-workspace
