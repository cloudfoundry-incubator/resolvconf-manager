#!/bin/bash

set -eu
set -o pipefail

export GOPATH="$(pwd)/resolvconf-manager-src"
export GOBIN="${GOPATH}/bin"
export PATH="${PATH}:${GOBIN}"
go get -v github.com/onsi/ginkgo/ginkgo &> /dev/null
go get ./... &> /dev/null
go install github.com/onsi/ginkgo/ginkgo &> /dev/null

pushd resolvconf-manager-src/
  ./scripts/test-unit.sh
popd
