#!/bin/bash

## Derived from https://github.com/goreleaser/get/blob/03c4bfde763b30bfe270892ab3ff74949d8e9351/get

set -euo pipefail

die () {
    echo >&2 "$@"
    exit 1
}

TAR_FILE="/tmp/ike.tar.gz"
RELEASES_URL="https://github.com/maistra/istio-workspace/releases"

show_help() {
  echo "get - downloads ike binary matching your operating system"
  echo " "
  echo "./get.sh [options]"
  echo " "
  echo "Options:"
  echo "-h, --help          shows brief help"
  echo "-v, --version       defines version specific version of the binary to download (defaults to latest)"
  echo "-d, --dir           target directory to which the binary is downloaded (defaults to random tmp dir in /tmp suffixed with ike-version)"
  echo "-n, --name          saves binary under specific name (defaults to ike)"
}

last_version() {
  curl -sL -o /dev/null -w %{url_effective} "$RELEASES_URL/latest" |
    rev |
    cut -f1 -d'/'|
    rev
}

download() {
  version=$1
  if [[ ${version} == "" ]]; then
    echo >&2 "Undefined version (pass using -v|--version). Please use semantic version. Read more about it here: https://semver.org/ \n\n"
    show_help
    exit 1
  fi

  url="$RELEASES_URL/download/$version/ike_${version:1}_$(uname -s)_$(uname -m).tar.gz"

  rm -f "$TAR_FILE"
  curl -fsLo "$TAR_FILE" "$url" || die "Unable to download $url"
}

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
              die "Please provide a version name"
            fi
            shift
            ;;
    --version*)
            version=`echo $1 | sed -e 's/^[^=]*=//g'`
            shift
            ;;
    -d)
            shift
            if test $# -gt 0; then
              dir=$1
            else
              die "Please provide a version"
            fi
            shift
            ;;
    --dir*)
            dir=`echo $1 | sed -e 's/^[^=]*=//g'`
            shift
            ;;
    -n)
            shift
            if test $# -gt 0; then
              binary_name=$1
            else
              die "Please provide a name"
            fi
            shift
            ;;
    --name*)
            binary_name=`echo $1 | sed -e 's/^[^=]*=//g'`
            shift
            ;;
    *)
            die "Unknown param $1"
            break
            ;;
  esac
done

binary_name=${binary_name:-"ike"}
version=${version:-"$(last_version)"}
test -z "$version" && {
  die "Unable to get ike version. You can still try to download it manually from $RELEASES_URL."
}
test -z ${dir+x} && {
  if [ "$(uname)" == "Darwin" ]; then
    dir="$(mktemp -d -t ike-${version}-)"
  else
    dir="$(mktemp -d --suffix=-ike-${version})"
  fi
}

download "${version}"
tar -C "$dir" -xzf "$TAR_FILE" ike
mv -n "$dir"/ike "$dir/$binary_name"

echo "Downloaded ike binary ($version) to $dir/$binary_name"
echo "Make sure it's on your \$PATH."
