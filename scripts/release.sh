#!/bin/bash

die () {
    echo >&2 "$@"
    exit 1
}

show_help() {
  echo "release - attempts to release new version of this project"
  echo " "
  echo "./release.sh [options]"
  echo " "
  echo "Options:"
  echo "-h, --help                shows brief help"
  echo "-v, --version=vx.y.yz     defines version for coming release. must be non-existing and following semantic rules"
  echo "-d, --dry-run             runs release process without doing actual push to git remote"
}

validate_version() {
  version=$1

  if [[ ${version} == "" ]]; then
    echo >&2 "Undefined version (pass using -v|--version). Please use semantic version. Read more about it here: https://semver.org/ \n\n"
    show_help
    exit 1
  fi

  tag_exists=$(git --no-pager tag --list | grep -c ${version})
  if [[ ${tag_exists} -ne 0 ]]; then
    die "Tag ${version} already exists!"
  fi

  # ensure defined version matches semver rules
  sem_ver_pattern="^[vV]?(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)(\\-[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?(\\+[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?$"
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

BASEDIR=$(git rev-parse --show-toplevel)
version=""
dry_run=false

if [[ "$#" -eq 0 ]]; then
  show_help
  exit 0
fi

while test $# -gt 0; do
  case "$1" in
    -h|--help)
            show_help
            exit 0
            ;;
    -v)
            shift
            if test $# -gt 0; then
              version=$1
            else
              die "Please provide tag name"
            fi
            shift
            ;;
    --version*)
            version=`echo $1 | sed -e 's/^[^=]*=//g'`
            shift
            ;;
    -d|--dry-run)
            shift
            dry_run=true
            shift
            ;;
    *)
            die "Unknown param $1"
            break
            ;;
  esac
done

validate_version ${version}

## Replace antora version for docs build
sed -i "/version:/c\version: ${version}" docs/antora.yml
sed -i "/^== Releases.*/a include::release_notes\/${version}.adoc[]\n" docs/modules/ROOT/pages/release_notes.adoc

git commit -am"release: ${version}"
git tag -a "${version}" -m"Automatically created release tag"

## Prepare next release iteration
sed -i "/version:/c\version: latest" docs/antora.yml
git commit -am"release: next iteration"

if ! ${dry_run}; then
  echo "Pushing changes to remote"
  git push && git push --tags
else
  echo "In dry-run mode, not pushing changes to remote"
fi
