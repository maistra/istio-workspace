#!/bin/bash

set -e

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

GIT_USER="${GIT_USER:-alien-ike}"
OWNER="${OWNER:-operator-framework}"
HUB_REPO_URL="${HUB_REPO_URL:-https://github.com/${OWNER}/community-operators.git}"
FORK="${FORK:-maistra}"
FORK_REPO_URL="${FORK_REPO_URL:-https://${GIT_USER}:${GITHUB_TOKEN}@github.com/${FORK}/community-operators.git}"

OPERATOR_NAME=istio-workspace-operator
OPERATOR_VERSION=${OPERATOR_VERSION:-0.0.5}
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

git add .
git commit -s -S -m"${TITLE}"

if [[ $runTests -ne 0 ]]; then
  echo "Running tests: $tests"

  cd "${TMP_DIR}"

  bash <(curl -sL https://cutt.ly/WhkV76k) \
  "$tests" \
  "${OPERATOR_HUB}/${OPERATOR_NAME}/${OPERATOR_VERSION}"
fi

if [[ -z $GITHUB_TOKEN ]]; then
  echo "Please provide GITHUB_TOKEN" && exit 1
fi

skipInDryRun git push fork "${BRANCH}"

PAYLOAD=$(mktemp)

jq -c -n \
  --arg msg "$(cat "${CUR_DIR}"/operatorhub-pr-template.md)" \
  --arg head "${FORK}:${BRANCH}" \
  --arg title "${TITLE}" \
   '{head: $head, base: "master", title: $title, body: $msg }' > "${PAYLOAD}"

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
