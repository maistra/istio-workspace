#!/bin/bash

die () {
    echo >&2 "$@"
    exit 1
}

show_help() {
  echo "release - attempts to release new version of this project"
  echo " "
  echo "./release.sh [flags|version]"
  echo " "
  echo "Options:"
  echo "-h, --help                shows brief help"
  echo "-v, --version=vx.y.yz     defines version for coming release. must be non-existing and following semantic rules"
  echo "                          this can also be passed as a first parameter to the script"
}

BASEDIR=$(git rev-parse --show-toplevel)
version=""

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
              die "Please provide version (according https://semver.org/)"
            fi
            shift
            ;;
    --version*)
            version=$(echo $1 | sed -e 's/^[^=]*=//g')
            shift
            ;;
    *)
            version=$(echo $1 | sed -e 's/^[^=]*=//g')
            shift
            ;;
  esac
done

## Check if tag exists
tag_exists=$(git --no-pager tag --list | grep -c "${version}")
if [[ ${tag_exists} -ne 0 ]]; then
  die "Tag ${version} already exists!"
fi

## Rebase the whole thing
branch=$(git rev-parse --abbrev-ref HEAD)
first_commit=$(git log master.."$branch" --oneline | tail -1 | cut -d' ' -f 1)
author=$(git show -s --format='%an <%ae>' "$first_commit")
title=$(git show -s --format='%s' "$first_commit")
git reset $(git merge-base master "$branch")

git add . && git commit --author="$author" --message="$title"

## Replace antora version for docs build
sed -i "/version:/c\version: ${version}" docs/antora.yml
sed -i "/^= Releases.*/a include::release_notes\/${version}.adoc[]\n" docs/modules/ROOT/pages/release_notes.adoc

git add . && git commit -F- <<EOF
release: ${version}

/tag ${version}
EOF

## Prepare next release iteration
sed -i "/version:/c\version: latest" docs/antora.yml
git commit -am"release: next iteration"

git push -f
