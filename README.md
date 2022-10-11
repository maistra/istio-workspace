[![Join Zulip chat](https://img.shields.io/badge/zulip-join_chat-brightgreen.svg)](https://istio-workspace.zulipchat.com/)
![Go Report Card](https://goreportcard.com/badge/github.com/maistra/istio-workspace)
![CircleCI](https://circleci.com/gh/maistra/istio-workspace.svg?style=svg)

# Istio Workspace
<!-- Following tag is extracted to operator description when building a bundle using `make bundle` -->
<!-- start::description[] -->
<!-- start::overview[] -->
Istio Workspace enables developers to:

* run one or more services locally during development but make it appear like it would be developed in the actual cluster
  * this way you can use your favourite development tools
  * have instant feedback loop
  * don't make your machine overheating trying to run the whole cluster locally
* safely preview incoming changes in pull requests without affecting regular users
* have confidence in testing or troubleshooting your services directly in the cluster

Istio Workspace supports testing on multi-user environment in an unobtrusive way.
It doesn’t really matter if it is QE cluster or actual production. We give you confidence that your changes won’t blow up the cluster, and your users won’t even notice a glitch.

You will need [Kubernetes](https://k8s.io) or [Openshift](https://openshift.com) cluster with [Istio](https://istio.io/) installed.

You can read more about our vision on [Red Hat Developer’s blog](https://developers.redhat.com/blog/2020/07/14/developing-and-testing-on-production-with-kubernetes-and-istio-workspace/) or ...

<!-- end::overview[] -->
## See it in action!

[![https://youtu.be/XTNVadUzMCc](https://img.youtube.com/vi/XTNVadUzMCc/hqdefault.jpg)](https://youtu.be/XTNVadUzMCc)

Watch the demo: ["How to develop on production: An introduction to Istio-Workspaces"](https://youtu.be/XTNVadUzMCc).

## Documentation

Head over to our [documentation page](https://istio-workspace-docs.netlify.com/) to learn the details!

<!-- end::description[] -->

We use amazing [Antora](https://antora.org/) project to build it and you should too!

## Install (in two easy steps)

Get latest `ike` binary through simple download script:

    curl -sL http://git.io/get-ike | bash

> TIP: You can also specify the version and directory before downloading `curl -sL http://git.io/get-ike | bash -s -- --version=v0.4.0 --dir=/usr/bin`

    get - downloads ike binary matching your operating system

    ./get.sh [options]

    Options:
    -h, --help          shows brief help
    -v, --version       defines version specific version of the binary to download (defaults to latest)
    -d, --dir           target directory to which the binary is downloaded (defaults to random tmp dir in /tmp suffixed with ike-version)
    -n, --name          saves binary under specific name (defaults to ike)

If you’re using Openshift you can install the `istio-workspace operator` via the Operator Hub in the web console.

If you’re on vanilla Kubernetes follow the instructions on [OperatorHub.io](https://operatorhub.io/operator/istio-workspace-operator) to see how you can install operators.

## Development Setup

<!-- start::dev-setup[] -->
Assuming that you have all the [Golang prerequisites](https://golang.org/doc/install) in place, clone the repository first:

    $ git clone https://github.com/maistra/istio-workspace $GOPATH/src/github.com/maistra/istio-workspace

From now on you are ready to hack - invoking `make` will check if required binaries are available and install if there are some missing.

Run `make help` to see what targets are available, but you will use `make` most often.

> NOTE: Have a look how [Go Version Manager](https://github.com/moovweb/gvm) can help you simplifying configuration
> and management of different versions of Go.

### Coding style

We follow standard Go coding conventions which we ensure using `goimports` during the build.

In addition, we provide `.editorconfig` file which is supported by [majority of the IDEs](https://editorconfig.org/#download). That should help avoid coding style mismatches.

<!-- end::dev-setup[] -->

## License

This project is licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/).
