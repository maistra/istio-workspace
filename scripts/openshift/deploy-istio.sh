#!/usr/bin/env bash

MAISTRA_VERSION=${MAISTRA_VERSION:-"0.12"}

if [ -z "$1" ]; then
    echo "-- Using default maistra version ${MAISTRA_VERSION}. You can override it by passing it as a first argument."
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
ISTIO_NS=${ISTIO_NS:-"istio-system"}
oc new-project "${ISTIO_NS}" || true
oc create -n "${ISTIO_NS}" -f deploy/istio/base-installation.yaml
echo "-- Waiting for ServiceMeshControlPlane to be ready ..."
until oc get ServiceMeshControlPlane -n "${ISTIO_NS}" -o go-template='{{range .items}}{{range .status.conditions}}{{.reason}}{{end}}{{end}}' | grep -m 1 "InstallSuccessful"; do : ; done
echo "... done"

echo "-- Adds admin user"
oc create user admin
oc adm policy add-cluster-role-to-user cluster-admin admin

echo "-- Expose docker registry"
oc expose service docker-registry -n default
