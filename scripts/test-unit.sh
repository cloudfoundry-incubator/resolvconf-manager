#!/bin/bash

set -eu
set -o pipefail

source_code=$(dirname $0)/..
source ${source_code}/.envrc
{
  go mod tidy && go mod verify
} &> /dev/null
pushd ${source_code}
  ginkgo -v --race --randomizeAllSpecs
popd
