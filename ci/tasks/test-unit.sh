#!/bin/bash

set -eux
set -o pipefail

export GOPATH="$(pwd)/resolvconf-manager-src"
export GOBIN="${GOPATH}/bin"
export PATH="${PATH}:${GOBIN}"
go get -v github.com/onsi/ginkgo/ginkgo
go get ./...
go install github.com/onsi/ginkgo/ginkgo

pushd resolvconf-manager-src/
  ./scripts/test-unit.sh
popd
