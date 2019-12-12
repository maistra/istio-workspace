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
              die "Please provide version (according https://semver.org/)"
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

./validate.sh ${version}

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
  echo "Don't forget to revert commits (i.e. git reset --hard HEAD~~) and delete the tag (git tag -d ${version})"
fi
