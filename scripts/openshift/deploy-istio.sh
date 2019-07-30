#!/usr/bin/env bash

MAISTRA_VERSION=${MAISTRA_VERSION:-"0.12"}

if [ -z "$1" ]; then
    echo "-- Using default maistra version ${MAISTRA_VERSION}. You can override it by passing it as a second argument."
else
    MAISTRA_VERSION=$1
fi

function wait_until_pod_started() {
    pod=$1
    ns=$2
    retries=${3:-20}
    count=0
    until oc -n ${ns} get pod $(oc get pods -n ${ns} | grep ${pod} | cut -d' ' -f 1) -o go-template='{{range .status.containerStatuses}}{{.state.running}}{{end}}' | grep -m 1 "startedAt" || [ ${count} -eq ${retries} ]; do
        sleep 5s
        ((count++))
    done
}

oc login -u system:admin

### Deploy istio-operator
oc new-project istio-operator || true
oc apply -n istio-operator -f https://raw.githubusercontent.com/Maistra/istio-operator/maistra-${MAISTRA_VERSION}/deploy/maistra-operator.yaml
wait_until_pod_started "istio-operator" "istio-operator"

sleep 5

###  Deploy service mesh
oc new-project istio-system || true
oc create -n istio-system -f deploy/istio/base-installation.yaml
echo "-- Waiting for ServiceMeshControlPlane to be ready ..."
until oc get ServiceMeshControlPlane -n istio-system -o go-template='{{range .items}}{{range .status.conditions}}{{.reason}}{{end}}{{end}}' | grep -m 1 "InstallSuccessful"; do : ; done
echo "... done"
