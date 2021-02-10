#!/bin/bash

set -e

GIT_USER="${GIT_USER:-alien-ike}"
CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
OWNER="${OWNER:-operator-framework}"
HUB_REPO_URL="${HUB_REPO_URL:-https://github.com/${OWNER}/community-operators.git}"
FORK="${FORK:-maistra}"
FORK_REPO_URL="${FORK_REPO_URL:-https://${GIT_USER}:${GITHUB_TOKEN}@github.com/${FORK}/community-operators.git}"
TEMP_FOLDER=$(mktemp -d)
OPERATOR_NAME=istio-workspace-operator
OPERATOR_VERSION=${OPERATOR_VERSION:-0.0.5} # TODO replace during release process?
BRANCH_NAME=${BRANCH_NAME:-"${OPERATOR_NAME}"-release-"${OPERATOR_VERSION}"}

die () {
    echo >&2 "$@"
    exit 1
}

if [[ -z $GITHUB_TOKEN ]]; then
  die "Please provide GITHUB_TOKEN"
fi

source "${CUR_DIR}"/validate_semver.sh
validate_semantic_versioning "v${OPERATOR_VERSION}"

git clone --depth 1 "${HUB_REPO_URL}" "${TEMP_FOLDER}"

cd "${TEMP_FOLDER}"
git remote add fork "${FORK_REPO_URL}"
git checkout -b "${BRANCH_NAME}"

mkdir -p community-operators/"${OPERATOR_NAME}"/"${OPERATOR_VERSION}"/
cp -a "${CUR_DIR}"/../../bundle/. community-operators/"${OPERATOR_NAME}"/"${OPERATOR_VERSION}"/

git add .
git commit -S -m"add ${OPERATOR_NAME} release ${OPERATOR_VERSION}"

git push fork "${BRANCH_NAME}"

TEMP_PAYLOAD=$(mktemp)

jq -c -n \
  --arg msg "$(cat "${CUR_DIR}"/operatorhub-pr-template.md)" \
  --arg head "${FORK}:${BRANCH_NAME}" \
  --arg title "add ${OPERATOR_NAME} release ${OPERATOR_VERSION}" \
   '{head: $head, base: "master", title: $title, body: $msg }' > "${TEMP_PAYLOAD}"

curl \
  -X POST \
  -H "Authorization: token ${GITHUB_TOKEN}" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/"${OWNER}"/community-operators/pulls \
   --data-binary "@${TEMP_PAYLOAD}"


