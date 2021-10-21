#!/bin/bash

set -euo pipefail

function replace_placeholders() {
  local path=$1
  local taskVersion=$2
  local image=$3
  local cmd=( "sed" )

  if [ -n "${4+x}" ]; then
      cmd+=( -i )
  fi
  cmd+=( -e "s/released-image/${image}/g" -e "s/current-version/${taskVersion}/g" "${path}")
  "${cmd[@]}"
}

case ${1-noop} in
    replace_placeholders) "$@"; exit;;
esac
