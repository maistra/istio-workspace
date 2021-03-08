#!/bin/bash

set -e

show_help() {
  echo "tektoncatalog - raises PR to Operator Hub"
  echo " "
  echo "./tektoncatalog.sh [options]"
  echo " "
  echo "Options:"
  echo "-h, --help              shows brief help"
  echo "-d, --dry-run           skips push to GH and PR"
}

dryRun=false

skipInDryRun() {
  if $dryRun; then echo "# $@"; fi
  if ! $dryRun; then "$@"; fi
}

while test $# -gt 0; do
  case "$1" in
  -h | --help)
    show_help
    exit 0
    ;;
  -d | --dry-run)
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

if ! command -v asciidoctor &>/dev/null; then
  echo "asciidoctor is required. Please install following packages:
  $ npm i -g asciidoctor @asciidoctor/core @asciidoctor/docbook-converter
  "
  exit 1
fi

if ! command -v pandoc &>/dev/null; then
  echo "pandoc is required. Please check installation guide at https://pandoc.org/installing.html"
  exit 1
fi

GIT_USER="${GIT_USER:-alien-ike}"
OWNER="${OWNER:-tektoncd}"
OWNER_REPO="${OWNER_REPO:-catalog}"
HUB_REPO_URL="${HUB_REPO_URL:-https://github.com/${OWNER}/${OWNER_REPO}.git}"
FORK="${FORK:-maistra}"
FORK_REPO="${FORK_REPO:-catalog}"
FORK_REPO_URL="${FORK_REPO_URL:-https://${GIT_USER}:${GITHUB_TOKEN}@github.com/${FORK}/${FORK_REPO}.git}"

OPERATOR_VERSION=${OPERATOR_VERSION:-0.0.5} # should be provided by Makefile target
TEKTON_HUB_PATH=${TEKTON_HUB_PATH:-task}

TITLE="Add istio-workspace release ${OPERATOR_VERSION} to ${TEKTON_HUB_PATH}"

CUR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
PROJECT_DIR="$(git rev-parse --show-toplevel)"
TASKS_DIR="${PROJECT_DIR}"/integration/tekton/tasks/
TMP_DIR=$(mktemp -d -t "tekton-${TEKTON_HUB_PATH}.XXXXXXXXXX")

# trap '{ rm -rf -- "$TMP_DIR"; }' EXIT

source "${CUR_DIR}"/validate_semver.sh

validate_semantic_versioning "v${OPERATOR_VERSION}"

##################################################################################
#### Prepare PR commit                                                        ####
##################################################################################

IKE_IMAGE="${IKE_DOCKER_REGISTRY}\/${IKE_DOCKER_REPOSITORY}\/${IKE_IMAGE_NAME}:${IKE_IMAGE_TAG}" # should be provided by Makefile target
BRANCH=${BRANCH:-"${TEKTON_HUB_PATH}/istio-workspace-${OPERATOR_VERSION}"}
ADOC_INCLUDE=$(LIB=$(mktemp) && wget -q -P "${LIB}" https://raw.githubusercontent.com/maistra/istio-workspace-docs-site/master/lib/include-shell.js && echo "${LIB}/include-shell")

git clone "${HUB_REPO_URL}" "${TMP_DIR}"
cd "${TMP_DIR}"
git remote add fork "${FORK_REPO_URL}"
git checkout -b "${BRANCH}"

COMMIT_MESSAGE=""
for taskName in "${TASKS_DIR}"/*; do
  taskName="${taskName##*/}"
  COMMIT_MESSAGE="${COMMIT_MESSAGE}
* Add Task ${taskName} release ${OPERATOR_VERSION}"

  mkdir -p "${TEKTON_HUB_PATH}/${taskName}/${OPERATOR_VERSION}"/
  cp -a "${TASKS_DIR}/${taskName}"/. "${TEKTON_HUB_PATH}/${taskName}/${OPERATOR_VERSION}"/

  pushd "${PROJECT_DIR}"
  asciidoctor --require @asciidoctor/docbook-converter --require "${ADOC_INCLUDE}" -a leveloffset=+1 --backend docbook -o - "${PROJECT_DIR}"/docs/modules/ROOT/pages/integration/tekton/tasks/"${taskName}".adoc | pandoc --wrap=preserve --from docbook --to gfm >"${TMP_DIR}/${TEKTON_HUB_PATH}/${taskName}/${OPERATOR_VERSION}"/README.md
  popd

  sed -i "s/released-image/${IKE_IMAGE}/g" "${TEKTON_HUB_PATH}/${taskName}/${OPERATOR_VERSION}/${taskName}.yaml"
done

COMMIT_MESSAGE="${COMMIT_MESSAGE}

$(
  cd "${PROJECT_DIR}"
  git log --pretty=format:%s $(git tag --sort=-committerdate | head -1)...$(git tag --sort=-committerdate | head -2 | awk '{split($0, tags, "\n")} END {print tags[1]}') integration/tekton | grep -s -v "release:" || true
)"

git add .
git commit -sS -m"${TITLE}

${COMMIT_MESSAGE}"

if [[ -z $GITHUB_TOKEN ]]; then
  echo "Please provide GITHUB_TOKEN" && exit 1
fi

skipInDryRun git push fork "${BRANCH}"

PAYLOAD=$(mktemp)

jq -c -n \
  --arg msg "$(awk -v msg="${COMMIT_MESSAGE}" '{gsub(/insert\-changes/,msg)}1' "${CUR_DIR}"/tektoncatalog-pr-template.md)" \
  --arg head "${FORK}:${BRANCH}" \
  --arg title "${TITLE}" \
  '{head: $head, base: "master", title: $title, body: $msg }' >"${PAYLOAD}"

if $dryRun; then
  echo -e "${PAYLOAD}\n------------------"
  jq . "${PAYLOAD}"
fi

skipInDryRun curl \
  -X POST \
  -H "Authorization: token ${GITHUB_TOKEN}" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/"${OWNER}/${OWNER_REPO}"/pulls \
  --data-binary "@${PAYLOAD}"
