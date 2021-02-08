#!/bin/bash

sem_ver_pattern="^[vV](0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)(\\-[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?(\\+[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?$"

die () {
    echo >&2 "$@"
    exit 1
}

validate_version() {
  version=$1

  if [[ ${version} == "" ]]; then
    die "Undefined version (pass using -v|--version). Please use semantic versioning https://semver.org/."
  fi

  # Ensure defined version matches semver rules
  if [[ ! "${version}" =~ $sem_ver_pattern ]]; then
    die "\`${version}\` you defined as a version does not match semantic versioning. Please make sure it conforms with https://semver.org/ and make sure it starts with v prefix."
  fi

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
