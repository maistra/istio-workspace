#!/bin/bash

set -e

GITHUB_USER="${GITHUB_USER:-alien-ike}"
CURR_FOLDER="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
OWNER="${OWNER:-operator-framework}"
HUB_REPO_URL="${HUB_REPO_URL:-https://github.com/${OWNER}/community-operators.git}"
FORK="${FORK:-maistra}"
FORK_REPO_URL="${FORK_REPO_URL:-https://${GITHUB_USER}:${GITHUB_TOKEN}@github.com/${FORK}/community-operators.git}"
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

# Validate version
# TODO: ./path doesn't work?
# TODO: validate failes due to existing v. Expected to be called before release, not after
#sh "${CURR_FOLDER}"/validate.sh "v${OPERATOR_VERSION}" --skip-ensure-release-notes

git clone "${HUB_REPO_URL}" "${TEMP_FOLDER}"

# make branch
cd "${TEMP_FOLDER}"

git remote add fork "${FORK_REPO_URL}"
git checkout -b "${BRANCH_NAME}"

mkdir -p community-operators/"${OPERATOR_NAME}"/"${OPERATOR_VERSION}"/
cp -R "${CURR_FOLDER}"/../../bundle/ community-operators/"${OPERATOR_NAME}"/"${OPERATOR_VERSION}"/

# commit - signed
git add .
git commit -S -m"add ${OPERATOR_NAME} release ${OPERATOR_VERSION}"

git push fork "${BRANCH_NAME}"


TEMP_PAYLOAD=$(mktemp)

jq -c -n \
  --arg msg "$(cat ${CURR_FOLDER}/operatorhub-pr-template.md)" \
  --arg head ${FORK}:${BRANCH_NAME} \
  --arg title "add ${OPERATOR_NAME} release ${OPERATOR_VERSION}" \
   '{head: $head, base: "master", title: $title, body: $msg }' > ${TEMP_PAYLOAD}

curl \
  -X POST \
  -H "Authorization: token ${GITHUB_TOKEN}" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/"${OWNER}"/community-operators/pulls \
   --data-binary "@${TEMP_PAYLOAD}"


