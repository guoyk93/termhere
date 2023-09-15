#!/bin/bash

set -eu

cd "$(dirname "${0}")"

rm -rf dist && mkdir dist

EXECUTABLE_NAME="termhere"

build() {
  rm -rf build && mkdir build
  GOOS=${1} GOARCH=${2} go build -o "build/${EXECUTABLE_NAME}${3}" ./cmd/${EXECUTABLE_NAME}
  tar -czvf "dist/${EXECUTABLE_NAME}-${1}-${2}.tar.gz" --exclude ".*" -C build "${EXECUTABLE_NAME}${3}"
  rm -rf build
}

build linux amd64 ""
build darwin amd64 ""
build darwin arm64 ""

build linux arm64 ""
build linux loong64 ""
build freebsd amd64 ""

cd dist

shasum -a 256 *.tar.gz > SHASUM256.txt
gpg-trezor-hi@guoyk.xyz -ab SHASUM256.txt
