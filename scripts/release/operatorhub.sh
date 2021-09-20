#!/bin/bash

set -euo pipefail

show_help() {
  echo "operatorhub - raises PR to Operator Hub"
  echo " "
  echo "./operatorhub.sh [options]"
  echo " "
  echo "Options:"
  echo "-h, --help              shows brief help"
  echo "-t, --test              runs operator-framework ansible tests (default: all. other supported values are: orange, kiwi or lemon)"
  echo "-d, --dry-run           skips push to GH and PR"
}

runTests=0
tests=all
dryRun=false

skipInDryRun() {
   if $dryRun; then echo "# $@";  fi
  if ! $dryRun; then "$@";  fi
}

while test $# -gt 0; do
  case "$1" in
    -h|--help)
            show_help
            exit 0
            ;;
    -t)
            shift
            runTests=1
            if test $# -gt 0; then
              tests=$1
            fi
            shift
            ;;
    --test*)
        runTests=1
        tests=$(echo $1 | sed -e 's/^[^=]*=//g')
        if [[ "$tests" == "--test" ]]; then
          tests=all
        fi
        shift
        ;;
    -d|--dry-run)
            dryRun=true
            shift
            ;;
    *)
            echo "Unknown param $1"
            exit 1
            break
            ;;
  esac
done

GITHUB_TOKEN="${GITHUB_TOKEN:-}"
GIT_USER="${GIT_USER:-alien-ike}"
OWNER="${OWNER:-operator-framework}"
HUB_REPO_URL="${HUB_REPO_URL:-https://github.com/${OWNER}/community-operators.git}"
HUB_BASE_BRANCH="${HUB_BASE_BRANCH:-master}"
FORK="${FORK:-maistra}"
FORK_REPO_URL="${FORK_REPO_URL:-https://${GIT_USER}:${GITHUB_TOKEN}@github.com/${FORK}/community-operators.git}"

OPERATOR_NAME=istio-workspace-operator
OPERATOR_VERSION=${OPERATOR_VERSION:-0.3.0}
OPERATOR_HUB=${OPERATOR_HUB:-community-operators}

BRANCH=${BRANCH:-"${OPERATOR_HUB}/${OPERATOR_NAME}-${OPERATOR_VERSION}"}
TITLE="Add ${OPERATOR_NAME} release ${OPERATOR_VERSION} to ${OPERATOR_HUB}"

CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
TMP_DIR=$(mktemp -d -t "${OPERATOR_HUB}-${OPERATOR_NAME}.XXXXXXXXXX")
trap '{ rm -rf -- "$TMP_DIR"; }' EXIT

source "${CUR_DIR}"/validate_semver.sh

validate_semantic_versioning "v${OPERATOR_VERSION}"

git clone "${HUB_REPO_URL}" "${TMP_DIR}"

cd "${TMP_DIR}"
git remote add fork "${FORK_REPO_URL}"
git checkout -b "${BRANCH}"

mkdir -p "${OPERATOR_HUB}/${OPERATOR_NAME}/${OPERATOR_VERSION}"/
cp -a "${CUR_DIR}"/../../bundle/. "${OPERATOR_HUB}/${OPERATOR_NAME}/${OPERATOR_VERSION}"/

skipInDryRun git add .
skipInDryRun git commit -s -S -m"${TITLE}"

if [[ $runTests -ne 0 ]]; then
  echo "Running tests: $tests"

  cd "${TMP_DIR}"

  ## can be removed after https://github.com/redhat-openshift-ecosystem/operator-test-playbooks/pull/244 is merged
  export OP_TEST_ANSIBLE_PULL_REPO="https://github.com/redhat-openshift-ecosystem/operator-test-playbooks"

  bash <(curl -sL https://cutt.ly/AEeucaw) \
  "$tests" \
  "${OPERATOR_HUB}/${OPERATOR_NAME}/${OPERATOR_VERSION}" > /tmp/test.out

  ## Until the script is fixed https://github.com/redhat-openshift-ecosystem/operator-test-playbooks/pull/247
  if tail -n 4 /tmp/test.out | grep "Failed with rc";
  then
    exit 1;
  fi # "Failed" was found in the logs
fi
if [[ ! $dryRun && -z $GITHUB_TOKEN ]]; then
  echo "Please provide GITHUB_TOKEN" && exit 1
fi

skipInDryRun git push fork "${BRANCH}"

PAYLOAD=$(mktemp)

jq -c -n \
  --arg msg "$(cat "${CUR_DIR}"/operatorhub-pr-template.md)" \
  --arg head "${FORK}:${BRANCH}" \
  --arg base "${HUB_BASE_BRANCH}" \
  --arg title "${TITLE}" \
   '{head: $head, base: $base, title: $title, body: $msg }' > "${PAYLOAD}"

if $dryRun; then
  echo -e "${PAYLOAD}\n------------------"
  jq . "${PAYLOAD}"
fi

skipInDryRun curl \
  -X POST \
  -H "Authorization: token ${GITHUB_TOKEN}" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/"${OWNER}"/community-operators/pulls \
   --data-binary "@${PAYLOAD}"
