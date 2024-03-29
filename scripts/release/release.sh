#!/bin/bash

set -euo pipefail

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
if [ $(git tag -l "$version") ]; then
  die "Version \`${version}\` already exists!"
fi

git reset $(git merge-base master HEAD)
git add -A
git commit -m "release: highlights of ${version}"

## Replace Antora version for docs build
sed -i "/version:/c\version: ${version}" docs/antora.yml
sed -i "/append:release_notes/a include::release_notes\/${version}.adoc[]\n" docs/modules/ROOT/pages/release_notes.adoc

## Bumps bundle
IKE_VERSION="${version}" make bundle

git add . && git commit -F- <<EOF
release: ${version}

/tag ${version}
EOF

## Prepare next release iteration
sed -i "/version:/c\version: latest" docs/antora.yml
IKE_VERSION="${version}-next" make bundle

git commit -am"release: next iteration" -m"/skip-e2e" -m"/skip-build"

git push -f
