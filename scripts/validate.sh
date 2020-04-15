#!/bin/bash

sem_ver_pattern="^[vV]?(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)(\\-[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?(\\+[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?$"

die () {
    echo >&2 "$@"
    exit 1
}

validate_version() {
  version=$1

  if [[ ${version} == "" ]]; then
    die "Undefined version (pass using -v|--version). Please use semantic versioning. Read more about it here: https://semver.org/"
  fi

  # ensure defined version matches semver rules
  if [[ ! "${version}" =~ $sem_ver_pattern ]]; then
    die "Version \`${version}\` you defined does not match semantic versioning. Read more about it here: https://semver.org/"
  fi

  # ensure release notes exist
  if [[ ! -f "docs/modules/ROOT/pages/release_notes/${version}.adoc" ]]; then
    die "Please create release notes in \`docs/modules/ROOT/pages/release_notes/${version}.adoc\`"
  fi
}

validate_version "$1"
