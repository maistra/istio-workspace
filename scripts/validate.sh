#!/bin/bash

sem_ver_pattern="^[vV]?(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)(\\-[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?(\\+[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?$"

die () {
    echo >&2 "$@"
    exit 1
}

validate_version() {
  version=$1

  if [[ ${version} == "" ]]; then
    die "Undefined version (pass using -v|--version). Please use semantic version. Read more about it here: https://semver.org/ \n\n"
  fi

  tag_exists=$(git --no-pager tag --list | grep -c ${version})
  if [[ ${tag_exists} -ne 0 ]]; then
    die "Tag ${version} already exists!"
  fi

  # ensure defined version matches semver rules
  if [[ ! "${version}" =~ $sem_ver_pattern ]]; then
    die "Version (${version}) you defined does not match semantic version. Read more about it here: https://semver.org/"
  fi

  # ensure release notes exist
  if [[ ! -f "docs/modules/ROOT/pages/release_notes/${version}.adoc" ]]; then
    die "Please create release notes in docs/modules/ROOT/pages/release_notes/${version}.adoc and submit it over a Pull Request."
  fi

  # ensure you are on master
  current_branch=$(git branch | grep \* | cut -d ' ' -f2)
  if [[ ${current_branch} != "master" ]]; then
    die "You are on ${current_branch} branch. Switch to master!"
  fi
}

validate_version "$1"
