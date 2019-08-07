#!/bin/bash

set -eu
set -o pipefail

source .envrc
{
  go mod tidy && go mod verify
} &> /dev/null
ginkgo -v --race --randomizeAllSpecs
