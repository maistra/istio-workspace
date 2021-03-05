#!/bin/bash

set -e

show_help() {
  echo "tektoncatalog - raises PR to Operator Hub"
  echo " "
  echo "./tektoncatalog.sh [options]"
  echo " "
  echo "Options:"
  echo "-h, --help              shows brief help"
  echo "-t, --test              runs operator-framework ansible tests (default: all. other supported values are: orange, kiwi or lemon)"
  echo "-d, --dry-run           skips push to GH and PR"
}

runTests=0
tests=all
dryRun=0

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


# For each sub folder in X
  # invoke asciidoctor README
  # sed container image version
  # commit, push, pr

GIT_USER="${GIT_USER:-alien-ike}"
OWNER="${OWNER:-tektoncd}"
OWNER_REPO="${OWNER_REPO:-catalog}"
HUB_REPO_URL="${HUB_REPO_URL:-https://github.com/${OWNER}/${OWNER_REPO}.git}"
FORK="${FORK:-maistra}"
FORK_REPO="${FORK_REPO:-catalog}"
FORK_REPO_URL="${FORK_REPO_URL:-https://${GIT_USER}:${GITHUB_TOKEN}@github.com/${FORK}/${FORK_REPO}.git}"

OPERATOR_VERSION=${OPERATOR_VERSION:-0.0.5}
OPERATOR_HUB=${OPERATOR_HUB:-task}

BRANCH=${BRANCH:-"${OPERATOR_HUB}/istio-workspace-${OPERATOR_VERSION}"}
TITLE="Add istio-workspace release ${OPERATOR_VERSION} to ${OPERATOR_HUB}"
COMMIT_MESSAGE="${TITLE}"

CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$( git rev-parse --show-toplevel )"
TMP_DIR=$(mktemp -d -t "${OPERATOR_HUB}-${OPERATOR_NAME}.XXXXXXXXXX")
#trap '{ rm -rf -- "$TMP_DIR"; }' EXIT

TASKS_DIR="${ROOT_DIR}"/integration/tekton/tasks/

source "${CUR_DIR}"/validate_semver.sh

validate_semantic_versioning "v${OPERATOR_VERSION}"

git clone "${HUB_REPO_URL}" "${TMP_DIR}"

ADOC_INCLUDE=$(LIB=$(mktemp) && wget -q -P $LIB https://raw.githubusercontent.com/maistra/istio-workspace-docs-site/master/lib/include-shell.js && echo "$LIB/include-shell")

cd "${TMP_DIR}"
git remote add fork "${FORK_REPO_URL}"
git checkout -b "${BRANCH}"

for OPERATOR_NAME in `ls ${TASKS_DIR}`
do

  COMMIT_MESSAGE="${COMMIT_MESSAGE}
* Add Task ${OPERATOR_NAME} release ${OPERATOR_VERSION}"

  mkdir -p "${OPERATOR_HUB}/${OPERATOR_NAME}/${OPERATOR_VERSION}"/
  cp -a "${TASKS_DIR}"${OPERATOR_NAME}/. "${OPERATOR_HUB}/${OPERATOR_NAME}/${OPERATOR_VERSION}"/

  pushd ${ROOT_DIR}
  asciidoctor --require @asciidoctor/docbook-converter --require ${ADOC_INCLUDE} -a leveloffset=+1 --backend docbook -o - ${ROOT_DIR}/docs/modules/ROOT/pages/integration/tekton/tasks/${OPERATOR_NAME}.adoc | pandoc --wrap=preserve --from docbook --to gfm > "${TMP_DIR}/${OPERATOR_HUB}/${OPERATOR_NAME}/${OPERATOR_VERSION}"/README.md
  popd

  sed -i "s/released-image/${IKE_DOCKER_REGISTRY}\/${IKE_DOCKER_REPOSITORY}\/${IKE_IMAGE_NAME}:${IKE_IMAGE_TAG}/g" "${OPERATOR_HUB}/${OPERATOR_NAME}/${OPERATOR_VERSION}"/${OPERATOR_NAME}.yaml
done

git add .
git commit -s -S -m"${COMMIT_MESSAGE}"

#if [[ $runTests -ne 0 ]]; then
#  echo "Running tests: $tests"
#
#  cd "${TMP_DIR}"
#
#  bash <(curl -sL https://cutt.ly/WhkV76k) \
#  "$tests" \
#  "${OPERATOR_HUB}/${OPERATOR_NAME}/${OPERATOR_VERSION}"
#fi

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
  --arg msg "$(cat "${CUR_DIR}"/tektoncatalog-pr-template.md)" \
  --arg head "${FORK}:${BRANCH}" \
  --arg title "${TITLE}" \
   '{head: $head, base: "master", title: $title, body: $msg }' > "${PAYLOAD}"

curl \
  -X POST \
  -H "Authorization: token ${GITHUB_TOKEN}" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/"${OWNER}"/${OWNER_REPO}/pulls \
   --data-binary "@${PAYLOAD}"
