#!/bin/bash

set -eux
set -o pipefail

source .envrc
go mod tidy && go mod verify
ginkgo -v --race --randomizeAllSpecs
