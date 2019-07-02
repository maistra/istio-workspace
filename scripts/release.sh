#!/bin/bash

validate_tag() {
  git_tag=$1

  tag_exists=$(git --no-pager tag --list | grep -c ${git_tag})
  if [[ ${tag_exists} -ne 0 ]]; then
    echo "Tag ${git_tag} already exists!" >&2
    exit 1
  fi

  sem_ver_pattern="^[vV]?(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)(\\-[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?(\\+[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?$"
  if [[ ! "${git_tag}" =~ $sem_ver_pattern ]]; then
    echo "Git tag you defined (${git_tag}) does not match semantic version. Read more about it here: https://semver.org/" >&2
    exit 1
  fi

  # ensure release notes exist
  if [[ ! -f "docs/modules/ROOT/pages/release_notes/${git_tag}.adoc" ]]; then
    echo "Please create release notes in docs/modules/ROOT/pages/release_notes/${git_tag}.adoc and submit it over a Pull Request." >&2
    exit 1
  fi

  # ensure you are on master
  current_branch=$(git branch | grep \* | cut -d ' ' -f2)
  if [[ ${current_branch} != "master" ]]; then
    echo "You are on ${current_branch} branch. Switch to master!" >&2
    exit 1
  fi

}

BASEDIR=$(git rev-parse --show-toplevel)

if [[ -z "$1" ]]; then
  echo "Please provide tag name" >&2
  exit 1
fi

git_tag=$1

validate_tag ${git_tag}

sed -i "/version:/c\version: ${git_tag}" docs/antora.yml
sed -i "/^== Releases.*/a include::release_notes\/${git_tag}.adoc[]\n" docs/modules/ROOT/pages/release_notes.adoc

git commit -am"release: ${git_tag}"
git tag -a "${git_tag}" -m"Automatically created release tag"

## Prepare next release iteration
sed -i "/version:/c\version: latest" docs/antora.yml
git commit -am"release: next iteration"

git push && git push --tags
