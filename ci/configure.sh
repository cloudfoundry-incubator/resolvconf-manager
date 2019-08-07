#!/usr/bin/env bash

dir=$(dirname $0)

until lpass status;do
  LPASS_DISABLE_PINENTRY=1 lpass ls a
  sleep 1
done

until fly -t production status;do
  fly -t production login
  sleep 1
done

fly -t production \
  sp -p resolvconf-manager \
  -l <(lpass show --notes 'resolvconf-manager pipeline vars') \
  -c $dir/pipeline.yml
