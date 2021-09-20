#!/usr/bin/env bash

set -euo pipefail 

protobuf_version="3.15.8"

get_arch() {
  case "$(uname -s)" in
    Darwin)
      echo 'osx-x86_64.zip'
      ;;
    Linux)
      echo 'linux-x86_64.zip'
      ;;
    CYGWIN*|MINGW32*|MSYS*)
      echo 'win64.zip'
      ;;
    *)
      echo 'other OS'
      ;;
  esac
}

url=https://github.com/google/protobuf/releases/download/v${protobuf_version}/protoc-${protobuf_version}-$(get_arch)

tmp_file=$(mktemp)
pwd=$(pwd)
wget -c -q --show-progress "${url}" -O "${tmp_file}"
unzip "${tmp_file}" 'bin/*' -d "${pwd}"
rm "${tmp_file}"
