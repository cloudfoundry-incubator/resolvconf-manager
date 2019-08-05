#!/usr/bin/env bash

set -ex
  lpass status
set +ex

dir=$(dirname $0)

fly -t ${CONCOURSE_TARGET:-production} \
  sp -p resolvconf-manager \
  -l <(lpass show --notes 'resolvconf-manager pipeline vars') \
  -c $dir/pipeline.yml
