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
            version=$(echo $1 | sed -e 's/^[^=]*=//g')
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

if ! /bin/bash "${BASEDIR}"/scripts/validate.sh "${version}"; then
  die
fi

## Check if tag exists
tag_exists=$(git --no-pager tag --list | grep -c "${version}")
if [[ ${tag_exists} -ne 0 ]]; then
  die "Tag ${version} already exists!"
fi

## Ensure you are on release branch
current_branch=$(git branch | grep "\*" | cut -d ' ' -f2)
if [[ ${current_branch} != "release_${version}" ]]; then
  die "You are on ${current_branch} branch. Switch to release_${version}!"
fi

## Generate changelog and append it to the highlights
ghc=$(curl -sL http://git.io/install-ghc | bash -s -- --path-only)
"${ghc}/ghc" generate -r maistra/istio-workspace --format adoc >> docs/modules/ROOT/pages/release_notes/"${version}".adoc

## Replace antora version for docs build
sed -i "/version:/c\version: ${version}" docs/antora.yml
sed -i "/^== Releases.*/a include::release_notes\/${version}.adoc[]\n" docs/modules/ROOT/pages/release_notes.adoc

git add . && git commit -am"release: ${version}"

## Prepare next release iteration
sed -i "/version:/c\version: latest" docs/antora.yml
git commit -am"release: next iteration"

if ! ${dry_run}; then
  echo "Pushing changes to remote"
  git push
else
  echo "Executed in dry-run mode, not pushing changes to remote."
  echo "Don't forget to revert commits (i.e. git reset --hard HEAD~~) and delete the tag if created (git tag -d ${version})."
fi
