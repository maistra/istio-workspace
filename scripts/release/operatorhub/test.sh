#!/bin/bash

# Script based on https://github.com/redhat-openshift-ecosystem/community-operators-pipeline/blob/41d2f3eb68eaf2df7cb02b2d2d10a5c7b45dfbd0/.github/workflows/operator_test.yaml

set -euo pipefail

show_help() {
  echo "test - runs Operator Hub tests"
  echo " "
  echo "./test.sh [options]"
  echo " "
  echo "Options:"
  echo "-h, --help              shows brief help"
  echo "-t, --test              runs operator-framework ansible tests (default: kiwi. supported values: all, orange, kiwi, lemon)"
}

tests=kiwi

while test $# -gt 0; do
  case "$1" in
    -h|--help)
        show_help
        exit 0
        ;;
    -t)
        shift
        if test $# -gt 0; then
          tests=$1
        fi
        shift
        ;;
    --test*)
        tests=$(echo $1 | sed -e 's/^[^=]*=//g')
        if [[ "$tests" == "--test" ]]; then
          tests=all
        fi
        shift
        ;;
    *)
        echo "Unknown param $1"
        exit 1
        break
        ;;
  esac
done

CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
BUNDLE_DIR="${CUR_DIR}"/../../../bundle

GITHUB_TOKEN="${GITHUB_TOKEN:-}"
GIT_USER="${GIT_USER:-alien-ike}"

OPERATOR_NAME=istio-workspace-operator
OPERATOR_VERSION=${OPERATOR_VERSION:-0.3.0}
OPERATOR_HUB=${OPERATOR_HUB:-"community-operators"}

TMP_DIR=$(mktemp -d -t "${OPERATOR_NAME}.XXXXXXXXXX")
trap '{ rm -rf -- "$TMP_DIR"; }' EXIT

#### We can keep old repository to run tests (upstream testing script still relies on that)
OWNER="${OWNER:-k8s-operatorhub}"
HUB_REPO_URL="${HUB_REPO_URL:-https://github.com/${OWNER}/${OPERATOR_HUB}.git}"
HUB_BASE_BRANCH="${HUB_BASE_BRANCH:-main}"

FORK="${FORK:-maistra}"
FORK_REPO_URL="${FORK_REPO_URL:-https://${GIT_USER}:${GITHUB_TOKEN}@github.com/${FORK}/community-operators.git}"

BRANCH=${BRANCH:-"${OPERATOR_HUB}/${OPERATOR_NAME}-${OPERATOR_VERSION}"}

source "${CUR_DIR}"/../validate_semver.sh

validate_semantic_versioning "v${OPERATOR_VERSION}"
echo "git clone ${HUB_REPO_URL} ${TMP_DIR}"
git clone "${HUB_REPO_URL}" "${TMP_DIR}"

cd "${TMP_DIR}"
git remote add fork "${FORK_REPO_URL}"
git checkout -b "${BRANCH}"

mkdir -p "operators/${OPERATOR_NAME}/${OPERATOR_VERSION}"/
cp -a "${BUNDLE_DIR}"/. "operators/${OPERATOR_NAME}/${OPERATOR_VERSION}"/

OPP_SCRIPT_URL="https://raw.githubusercontent.com/redhat-openshift-ecosystem/community-operators-pipeline/ci/latest/ci/scripts/opp.sh"
OPP_SCRIPT_ENV_URL="https://raw.githubusercontent.com/redhat-openshift-ecosystem/community-operators-pipeline/ci/latest/ci/scripts/opp-env.sh"
OPP_SCRIPT_ENV_OPRT_URL="https://raw.githubusercontent.com/redhat-openshift-ecosystem/community-operators-pipeline/ci/latest/ci/scripts/opp-oprt.sh"
OPP_IMAGE="quay.io/operator_testing/operator-test-playbooks:latest"
OPP_ANSIBLE_PULL_REPO="https://github.com/redhat-openshift-ecosystem/operator-test-playbooks/"
OPP_ANSIBLE_PULL_BRANCH="upstream-community"
OPP_THIS_REPO_BASE="https://github.com"
OPP_THIS_REPO="${FORK}/${OPERATOR_HUB}"
OPP_THIS_BRANCH="main"
OPP_RELEASE_BUNDLE_REGISTRY="quay.io"
OPP_RELEASE_BUNDLE_ORGANIZATION="operatorhubio"
OPP_RELEASE_INDEX_REGISTRY="quay.io"
OPP_RELEASE_INDEX_ORGANIZATION="operatorhubio"
OPP_RELEASE_INDEX_NAME="catalog"
OPP_MIRROR_INDEX_MULTIARCH_BASE="registry.redhat.io/openshift4/ose-operator-registry:v4.9"
OPP_MIRROR_INDEX_MULTIARCH_POSTFIX="s"
KIND_VERSION="v0.14.0"
KIND_KUBE_VERSION="v1.21.1"
OPP_PRODUCTION_TYPE="k8s"


echo "Running tests: $tests"
cd "${TMP_DIR}"

bash <(curl -sL $OPP_SCRIPT_URL) \
  "$tests" \
  "operators/${OPERATOR_NAME}/${OPERATOR_VERSION}"

## Until the script is fixed https://github.com/redhat-openshift-ecosystem/operator-test-playbooks/pull/247
if tail -n 4 /tmp/test.out | grep "Failed with rc";
then
  exit 1;
fi # "Failed" was found in the logs
