#!/bin/bash

set -e

show_help() {
  echo "operatorhub - raises PR to Operator Hub"
  echo " "
  echo "./operatorhub.sh [options]"
  echo " "
  echo "Options:"
  echo "-h, --help              shows brief help"
  echo "-t, --test              runs operator-framework ansible tests"
  echo "-d, --dry-run           skips push to GH and PR"
}

runTests=0
dryRun=0

while test $# -gt 0; do
  case "$1" in
    -h|--help)
            show_help
            exit 0
            ;;
    -t|--test)
            runTests=1
            shift
            ;;
    -d|--dry-run)
            dryRun=1
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
TMP_DIR=$(mktemp -d)

source "${CUR_DIR}"/validate_semver.sh

GIT_USER="${GIT_USER:-alien-ike}"
OWNER="${OWNER:-operator-framework}"
HUB_REPO_URL="${HUB_REPO_URL:-https://github.com/${OWNER}/community-operators.git}"
FORK="${FORK:-maistra}"
FORK_REPO_URL="${FORK_REPO_URL:-https://${GIT_USER}:${GITHUB_TOKEN}@github.com/${FORK}/community-operators.git}"

OPERATOR_NAME=istio-workspace-operator
OPERATOR_VERSION=${OPERATOR_VERSION:-0.0.5}
OPERATOR_HUB=${OPERATOR_HUB:-community-operators}

BRANCH=${BRANCH:-"${OPERATOR_HUB}/${OPERATOR_NAME}"-release-"${OPERATOR_VERSION}"}
TITLE="Add ${OPERATOR_NAME} release ${OPERATOR_VERSION} to ${OPERATOR_HUB}"

validate_semantic_versioning "v${OPERATOR_VERSION}"

git clone --depth 1 "${HUB_REPO_URL}" "${TMP_DIR}"

cd "${TMP_DIR}"
git remote add fork "${FORK_REPO_URL}"
git checkout -b "${BRANCH}"

mkdir -p "${OPERATOR_HUB}/${OPERATOR_NAME}/${OPERATOR_VERSION}"/
cp -a "${CUR_DIR}"/../../bundle/. "${OPERATOR_HUB}/${OPERATOR_NAME}/${OPERATOR_VERSION}"/

git add .
git commit -S -m"${TITLE}"

if [[ $runTests -ne 0 ]]; then
  echo "Running tests"

  cd "${TMP_DIR}"

  bash <(curl -sL https://cutt.ly/WhkV76k) \
  all \
  "${OPERATOR_HUB}/${OPERATOR_NAME}/${OPERATOR_VERSION}"
fi

if [[ $dryRun -ne 0 ]]; then
    echo "skips pushing to Git Hub"
    exit 0
fi

if [[ -z $GITHUB_TOKEN ]]; then
  echo "Please provide GITHUB_TOKEN" && exit 1
fi

git push fork "${BRANCH}"

PAYLOAD=$(mktemp)

jq -c -n \
  --arg msg "$(cat "${CUR_DIR}"/operatorhub-pr-template.md)" \
  --arg head "${FORK}:${BRANCH}" \
  --arg title "${TITLE}" \
   '{head: $head, base: "master", title: $title, body: $msg }' > "${PAYLOAD}"

curl \
  -X POST \
  -H "Authorization: token ${GITHUB_TOKEN}" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/"${OWNER}"/community-operators/pulls \
   --data-binary "@${PAYLOAD}"
