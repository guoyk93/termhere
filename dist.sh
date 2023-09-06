#!/bin/bash

set -eu

cd "$(dirname "${0}")"

rm -rf dist && mkdir dist

build() {
  rm -rf build && mkdir build
  GOOS=${1} GOARCH=${2} go build -o build/termhere ./cmd/termhere
  tar -czvf "dist/termhere-${1}-${2}.tar.gz" --exclude ".*" -C build termhere
  rm -rf build
}

build darwin arm64
build darwin amd64
build linux arm64
build linux amd64
build linux loong64
build linux riscv64
