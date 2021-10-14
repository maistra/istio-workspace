#!/usr/bin/env bash

set -euo pipefail

if ! command -v pandoc &> /dev/null
then
    echo "pandoc could not be found. installing..."
    PANDOC_VERSION=2.11.4
    cd /tmp
    wget "https://github.com/jgm/pandoc/releases/download/${PANDOC_VERSION}/pandoc-${PANDOC_VERSION}-linux-amd64.tar.gz" -O "pandoc.tar.gz"
    tar xzfv pandoc.tar.gz
    sudo mv "${PWD}"/pandoc-${PANDOC_VERSION}/bin/pandoc /usr/local/bin/
    cd -
fi

function update_sections() {
  name="$1"
  path="$2"
  awk '/start::'"${name}"'/{flag=1;next}/end::'"${name}"'/{flag=0}flag' README.md | pandoc --wrap=preserve --from gfm --to asciidoc > /tmp/"${name}"
  sed -i -ne '/start:'"${name}"'/{p;r /tmp/'"${name}"'' -e ':a;n;/end:'"${name}"'/!ba};p' "${path}"
}

update_sections "overview" "docs/modules/ROOT/pages/index.adoc"
update_sections "dev-setup" "docs/modules/ROOT/pages/contribution_guide.adoc"
