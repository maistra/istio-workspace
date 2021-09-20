#!/bin/bash

set -euo pipefail

CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

source "${CUR_DIR}"/validate_semver.sh

die () {
    echo >&2 "$@"
    exit 1
}

validate_version() {
  version=$1

  validate_semantic_versioning "$1"

  ## Check if tag exists
  tag_exists=$(git --no-pager tag --list | grep -c "${version}")
  if [[ ${tag_exists} -ne 0 ]]; then
    die "Version \`${version}\` already exists!"
  fi

}

ensure_release_notes() {
  version=$1

  # Ensure release notes exist
  if [[ ! -f "docs/modules/ROOT/pages/release_notes/${version}.adoc" ]]; then
    die "It seems you want to release \`${version}\`. Please create release highlights in \`docs/modules/ROOT/pages/release_notes/${version}.adoc\`."
  fi
}

validate_version "$1"

if [[ "$#" -eq 1 ]]; then
  ensure_release_notes "$1"
fi
