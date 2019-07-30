#!/usr/bin/env bash

IKE_CLUSTER_DIR=${IKE_CLUSTER_DIR:-"~/openshift.local.cluster"}

if [ -z "$1" ]; then
    [ -d "${IKE_CLUSTER_DIR}" ] || mkdir -p "${IKE_CLUSTER_DIR}"
else
    IKE_CLUSTER_DIR=$1
fi

echo "-- Starting cluster in ${IKE_CLUSTER_DIR}. You can override it by passing directory as a first argument or by setting environment variable IKE_CLUSTER_DIR."

PATCH=$(cat <<EOF
admissionConfig:
  pluginConfig:
    MutatingAdmissionWebhook:
      configuration:
        apiVersion: apiserver.config.k8s.io/v1alpha1
        kubeConfigFile: /dev/null
        kind: WebhookAdmission
    ValidatingAdmissionWebhook:
      configuration:
        apiVersion: apiserver.config.k8s.io/v1alpha1
        kubeConfigFile: /dev/null
        kind: WebhookAdmission
EOF
)

### Master config patch
MASTER_CONFIG_PATH="${IKE_CLUSTER_DIR}/kube-apiserver/master-config.yaml"
oc cluster up --base-dir "${IKE_CLUSTER_DIR}" --write-config
cp "${MASTER_CONFIG_PATH}" "${MASTER_CONFIG_PATH}".backup
cp "${MASTER_CONFIG_PATH}" "${MASTER_CONFIG_PATH}".patch
oc ex config patch "${MASTER_CONFIG_PATH}".patch -p "${PATCH}" > "${MASTER_CONFIG_PATH}"

oc cluster up --base-dir "${IKE_CLUSTER_DIR}"
