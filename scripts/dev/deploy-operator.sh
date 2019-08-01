#!/usr/bin/env bash

export IKE_DOCKER_REGISTRY=docker-registry-default.127.0.0.1.nip.io:80
export IKE_DOCKER_REPOSITORY=istio-workspace-operator
export TELEPRESENCE_VERSION=$(telepresence --version)
export IKE_IMAGE_TAG="$(make version)"

oc new-project istio-workspace-operator || true

docker login -u $(oc whoami) -p $(oc whoami -t) $(echo $IKE_DOCKER_REGISTRY)
make docker-build docker-push
export IKE_DOCKER_REGISTRY=172.30.1.1:5000
ike install-operator
