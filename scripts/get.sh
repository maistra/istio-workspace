#!/bin/bash

## Derived from https://github.com/goreleaser/get/blob/03c4bfde763b30bfe270892ab3ff74949d8e9351/get

set -e

die () {
    echo >&2 "$@"
    exit 1
}

TAR_FILE="/tmp/ike.tar.gz"
RELEASES_URL="https://github.com/maistra/istio-workspace/releases"

last_version() {
  curl -sL -o /dev/null -w %{url_effective} "$RELEASES_URL/latest" |
    rev |
    cut -f1 -d'/'|
    rev
}

download() {
  rm -f "$TAR_FILE"
  curl -s -L -o "$TAR_FILE" \
    "$RELEASES_URL/download/$VERSION/ike_${VERSION:1}_$(uname -s)_$(uname -m).tar.gz"
}

test -z "$VERSION" && VERSION="$(last_version)"
test -z "$VERSION" && {
  die "Unable to get ike version. You can still try to download manually from $RELEASES_URL"
}
test -z "$TMPDIR" && TMPDIR="$(mktemp -d --suffix=-ike-${VERSION})"

download
tar -xf "$TAR_FILE" -C "$TMPDIR"

echo "Downloaded ike binary to $TMPDIR"
echo -e "You can add it to your path by typing following in your terminal:\n$ export PATH=\"\$PATH:$TMPDIR\""
