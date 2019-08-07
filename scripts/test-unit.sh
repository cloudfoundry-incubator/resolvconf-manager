#!/bin/bash

set -eux
set -o pipefail

source_code=$(dirname $0)/..
source ${source_code}/.envrc
pushd ${source_code}
  ginkgo -v --race --randomizeAllSpecs
popd
