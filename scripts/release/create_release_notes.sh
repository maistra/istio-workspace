#!/bin/bash

set -euo pipefail

CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PROJECT_ROOT_DIR=$(git rev-parse --show-toplevel)

die() {
  echo >&2 "$@"
  exit 1
}

dryRun=false

skipInDryRun() {
  if $dryRun; then echo "# $@";  fi
  if ! $dryRun; then "$@";  fi
}

show_help() {
  echo "create_release_notes - prepares commits for release notes process"
  echo " "
  echo "./create_release_notes.sh [options]"
  echo " "
  echo "Options:"
  echo "-h, --help        shows brief help"
  echo "-v, --version     targeted version release (required, must be semver compliant)"
  echo "-d, --dry-run     skips push to GH and PR"
}

version="null"

while test $# -gt 0; do
  case "$1" in
    -h|--help)
            show_help
            exit 0
            ;;
    -d|--dry-run)
            dryRun=true
            shift
            ;;
    -v)
            shift
            if test $# -gt 0; then
              version=$1
            else
              die "Please provide version (according https://semver.org/)"
            fi
            shift
            ;;
    --version*)
            version=$(echo $1 | sed -e 's/^[^=]*=//g')
            shift
            ;;
    *)
            echo "Unknown param $1"
            exit 1
            break
            ;;
  esac
done

if [[ -z $version || $version == "null" || $version == "x" ]]; then
  die "Please specify version you are targeting for the release."
fi

source "${CUR_DIR}"/validate_semver.sh
validate_semantic_versioning "${version}" --skip-release-notes-check
skipInDryRun git checkout -b release_"${version}"
cp "${PROJECT_ROOT_DIR}"/docs/modules/ROOT/pages/release_notes/release_notes_template.adoc "${PROJECT_ROOT_DIR}"/docs/modules/ROOT/pages/release_notes/"${version}".adoc
sed -i -e "s/vX.Y.Z/${version}/" docs/modules/ROOT/pages/release_notes/"${version}".adoc
skipInDryRun git add .
skipInDryRun git commit -m "release: highlights of ${version}" -m "/skip-e2e" -m "/skip-build"
skipInDryRun git show HEAD
